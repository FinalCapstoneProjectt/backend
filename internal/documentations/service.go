package documentations

import (
	"backend/internal/domain"
	"backend/internal/files"
	"errors"
	"path/filepath"
	"strings"

	"mime/multipart"
	"time"
)

type Service struct {
	repo     Repository
	uploader *files.Uploader
}

func NewService(r Repository, u *files.Uploader) *Service {
	return &Service{repo: r, uploader: u}
}

func (s *Service) SubmitDoc(projectID, userID uint, docType, url string, file *multipart.FileHeader) (*domain.ProjectDocumentation, error) {
	// 1. Check if THIS SPECIFIC document type already exists for this project
	existing, _ := s.repo.GetByType(projectID, docType)
	if existing != nil && existing.ID != 0 {
		return nil, errors.New("this specific document/link already exists. Delete it first to re-upload")
	}

	finalURL := url

	// 2. Handle physical file validation and upload
	if file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		
		// 🔒 STRICT EXTENSION VALIDATION
		if docType == "final_report" && ext != ".pdf" {
			return nil, errors.New("invalid file type: Final Report must be a PDF")
		}
		if docType == "presentation" && ext != ".ppt" && ext != ".pptx" {
			return nil, errors.New("invalid file type: Presentation must be PPT or PPTX")
		}

		path, err := s.uploader.SaveFile(file, "project_docs")
		if err != nil { return nil, err }
		finalURL = path
	}

	doc := &domain.ProjectDocumentation{
		ProjectID:    projectID,
		DocumentType: docType, // 'final_report', 'presentation', 'code_link', 'deployed_link'
		URL:          finalURL,
		Status:       "pending",
		SubmittedBy:  userID,
		SubmittedAt:  time.Now(),
	}

	if err := s.repo.Create(doc); err != nil { return nil, err }
	return doc, nil
}

func (s *Service) DeleteDoc(docID, userID uint) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil { return errors.New("document not found") }

	// 🔒 RULE: Only Pending can be unlinked/deleted
	if doc.Status != "pending" {
		return errors.New("cannot unlink an approved document. Contact your advisor")
	}

	// 🔒 Check if it's a physical file or just a link
	isPhysicalFile := doc.DocumentType == "final_report" || doc.DocumentType == "presentation"
	
	if isPhysicalFile {
		// Remove from hard drive
		_ = s.uploader.DeleteFile(doc.URL)
	}

	// Always remove from Database to allow student to re-submit
	return s.repo.Delete(docID)
}

func (s *Service) ReviewDoc(docID, reviewerID uint, status string, comment string) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil { return err }

	doc.Status = status
	doc.ReviewComment = comment
	doc.ReviewedBy = reviewerID
	doc.ReviewedAt = time.Now()

	// 🔒 RULE: If Rejected, delete the physical file
	if status == "rejected" {
		_ = s.uploader.DeleteFile(doc.URL)
		return s.repo.Delete(docID) // Remove from DB too as per your request
	}

	return s.repo.Update(doc)
}

func (s *Service) GetDocs(projectID uint) ([]domain.ProjectDocumentation, error) {
	return s.repo.GetByProjectID(projectID)
}