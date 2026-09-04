package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
	"poem-backend/pkg/convert"
	"poem-backend/pkg/response"
)

type AdminAuthorService struct {
	authorRepo *repository.AuthorRepository
}

func NewAdminAuthorService(authorRepo *repository.AuthorRepository) *AdminAuthorService {
	return &AdminAuthorService{authorRepo: authorRepo}
}

// List 分页获取作者列表
func (s *AdminAuthorService) List(ctx context.Context, page, pageSize int, keyword string) (*response.PageData[adminmodel.AdminAuthorResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	authors, total, err := s.authorRepo.List(ctx, page, pageSize, keyword)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询作者列表失败: %v", err)}
	}

	items := make([]adminmodel.AdminAuthorResponse, 0, len(authors))
	for _, a := range authors {
		poemCount, _ := s.authorRepo.GetPoemCount(ctx, a.ID)
		items = append(items, toAdminAuthorResponse(a, poemCount))
	}
	return &response.PageData[adminmodel.AdminAuthorResponse]{Items: items, Total: total}, nil
}

// GetByID 获取作者详情
func (s *AdminAuthorService) GetByID(ctx context.Context, id int64) (*adminmodel.AdminAuthorResponse, error) {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "author not found", Detail: fmt.Sprintf("作者不存在: id=%d", id)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询作者失败: id=%d, error=%v", id, err)}
	}
	poemCount, _ := s.authorRepo.GetPoemCount(ctx, author.ID)
	resp := toAdminAuthorResponse(*author, poemCount)
	return &resp, nil
}

// Create 创建作者
func (s *AdminAuthorService) Create(ctx context.Context, req *adminmodel.AdminAuthorCreateRequest) (*adminmodel.AdminAuthorResponse, error) {
	// 处理繁体：如果未提供则尝试从简体转换
	nameTraditional := req.NameTraditional
	if nameTraditional == "" && req.Name != "" {
		nameTraditional = convert.MustSimplifiedToTraditional(req.Name)
	}

	author := &model.Author{
		Name:            req.Name,
		NameTraditional: nameTraditional,
		Dynasty:         req.Dynasty,
		Biography:       req.Biography,
	}
	if author.Dynasty == "" {
		author.Dynasty = "未知"
	}

	if err := s.authorRepo.Create(ctx, author); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建作者失败: %v", err)}
	}

	resp := toAdminAuthorResponse(*author, 0)
	return &resp, nil
}

// Update 更新作者
func (s *AdminAuthorService) Update(ctx context.Context, id int64, req *adminmodel.AdminAuthorUpdateRequest) error {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fuego.NotFoundError{Title: "author not found", Detail: fmt.Sprintf("作者不存在: id=%d", id)}
		}
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询作者失败: id=%d, error=%v", id, err)}
	}

	author.Name = req.Name
	author.NameTraditional = req.NameTraditional
	author.Dynasty = req.Dynasty
	author.Biography = req.Biography
	if author.Dynasty == "" {
		author.Dynasty = "未知"
	}

	if err := s.authorRepo.Update(ctx, author); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新作者失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// Delete 删除作者
func (s *AdminAuthorService) Delete(ctx context.Context, id int64) error {
	if err := s.authorRepo.Delete(ctx, id); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除作者失败: id=%d, error=%v", id, err)}
	}
	return nil
}

// SearchOptions 搜索作者（用于下拉框）
func (s *AdminAuthorService) SearchOptions(ctx context.Context, keyword string) ([]adminmodel.AdminAuthorOptionResponse, error) {
	authors, err := s.authorRepo.SearchByKeyword(ctx, keyword, 20)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("搜索作者失败: %v", err)}
	}

	options := make([]adminmodel.AdminAuthorOptionResponse, 0, len(authors))
	for _, a := range authors {
		options = append(options, adminmodel.AdminAuthorOptionResponse{
			ID:      a.ID,
			Name:    a.Name,
			Dynasty: a.Dynasty,
		})
	}
	return options, nil
}

// BatchMatchPoems 批量匹配诗歌关联作者
func (s *AdminAuthorService) BatchMatchPoems(ctx context.Context, poetryIDs []int64) (*adminmodel.AdminAuthorBatchMatchResponse, error) {
	matched, unmatched, err := s.authorRepo.BatchMatchPoems(ctx, poetryIDs)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("批量匹配诗歌失败: %v", err)}
	}
	return &adminmodel.AdminAuthorBatchMatchResponse{
		Total:     int64(len(poetryIDs)),
		Matched:   matched,
		Unmatched: unmatched,
	}, nil
}

// GenerateAuthorsFromPoems 从诗歌中提取不重复的作者名，自动创建作者记录
// 返回统计信息：唯一作者数、新建数、跳过数
func (s *AdminAuthorService) GenerateAuthorsFromPoems(ctx context.Context) (*adminmodel.AdminToolGenerateAuthorsResponse, error) {
	result, err := s.authorRepo.GenerateAuthorsFromPoems(ctx)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("从诗歌提取作者失败: %v", err)}
	}
	return result, nil
}

// AuthorDedupScan 扫描重复作者组
func (s *AdminAuthorService) AuthorDedupScan(ctx context.Context, matchBy string) (*adminmodel.AdminToolAuthorDedupScanResponse, error) {
	result, err := s.authorRepo.AuthorDedupScan(ctx, matchBy)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("扫描重复作者失败: %v", err)}
	}
	return result, nil
}

// AuthorDedupMerge 合并重复作者
func (s *AdminAuthorService) AuthorDedupMerge(ctx context.Context, keepID int64, mergeIDs []int64) (*adminmodel.AdminToolAuthorDedupMergeResponse, error) {
	result, err := s.authorRepo.AuthorDedupMerge(ctx, keepID, mergeIDs)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("合并重复作者失败: %v", err)}
	}
	return result, nil
}

// CleanupAuthorNames 清理 name = name_traditional 的记录
func (s *AdminAuthorService) CleanupAuthorNames(ctx context.Context) (int64, string, error) {
	cleaned, err := s.authorRepo.CleanupAuthorNames(ctx)
	if err != nil {
		return 0, "", fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("清理作者繁体名失败: %v", err)}
	}
	return cleaned, fmt.Sprintf("清理完成：已将 %d 个作者的繁体名清空（与简体相同）", cleaned), nil
}

// EnsureAuthorNamesSimplified 确保 authors 表的 name 字段为简体字
func (s *AdminAuthorService) EnsureAuthorNamesSimplified(ctx context.Context) (int64, string, error) {
	processed, err := s.authorRepo.EnsureAuthorNamesSimplified(ctx)
	if err != nil {
		return 0, "", fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("作者姓名转简体失败: %v", err)}
	}
	return processed, fmt.Sprintf("处理完成：已将 %d 个作者的姓名转为简体（原值保留为繁体）", processed), nil
}

// toAdminAuthorResponse 转换 Author 为 AdminAuthorResponse
func toAdminAuthorResponse(a model.Author, poemCount int64) adminmodel.AdminAuthorResponse {
	return adminmodel.AdminAuthorResponse{
		ID:              a.ID,
		Name:            a.Name,
		NameTraditional: a.NameTraditional,
		Dynasty:         a.Dynasty,
		Biography:       a.Biography,
		PoemCount:       poemCount,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}
