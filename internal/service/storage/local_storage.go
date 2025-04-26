package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/service" // 引入 service 包以引用接口
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// LocalStorage 实现了 FileStorage 接口，使用本地文件系统存储文件。
type LocalStorage struct {
	uploadDir string // 从配置中获取的基础上传目录
}

// NewLocalStorage 创建一个新的 LocalStorage 实例。
// 它会检查并创建配置中指定的上传目录（如果不存在）。
func NewLocalStorage(cfg *config.Config) (service.FileStorage, error) {
	uploadDir := cfg.UploadDir
	logger.Info("初始化本地文件存储...", "upload_dir", uploadDir)

	// 检查并创建基础上传目录
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("创建上传目录失败", "error", err, "path", uploadDir)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法初始化文件存储目录")
	}

	return &LocalStorage{uploadDir: uploadDir}, nil
}

// SaveFile 将文件保存到本地磁盘。
// 文件将保存在 uploadDir/userID/uuid_filename 路径下。
func (ls *LocalStorage) SaveFile(ctx context.Context, userID string, filename string, fileData io.Reader) (storedPath string, err error) {
	if userID == "" {
		return "", apperr.New(apperr.CodeInvalidArgument, "用户 ID 不能为空")
	}

	// 创建用户特定的子目录
	userDir := filepath.Join(ls.uploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		logger.ErrorContext(ctx, "创建用户上传子目录失败", "error", err, "path", userDir)
		return "", apperr.Wrap(err, apperr.CodeInternal, "无法创建用户目录")
	}

	// 生成一个唯一的文件名以避免冲突，但保留原始扩展名
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	storedPath = filepath.Join(userDir, uniqueFilename)

	logger.InfoContext(ctx, "准备保存文件到本地", "user_id", userID, "original_filename", filename, "stored_path", storedPath)

	// 创建目标文件
	dst, err := os.Create(storedPath)
	if err != nil {
		logger.ErrorContext(ctx, "创建目标文件失败", "error", err, "path", storedPath)
		return "", apperr.Wrap(err, apperr.CodeInternal, "无法创建目标文件")
	}
	defer func() {
		closeErr := dst.Close()
		if err == nil && closeErr != nil { // 如果主操作成功，但关闭失败，则报告关闭错误
			err = apperr.Wrap(closeErr, apperr.CodeInternal, "关闭文件时出错")
			logger.ErrorContext(ctx, "关闭保存的文件时出错", "error", closeErr, "path", storedPath)
		} else if closeErr != nil { // 如果主操作失败，并且关闭也失败，记录关闭错误但返回主错误
			logger.WarnContext(ctx, "关闭文件时出错（已存在其他错误）", "error", closeErr, "path", storedPath, "original_error", err)
		}
	}()

	// 将上传的文件内容复制到目标文件
	bytesWritten, err := io.Copy(dst, fileData)
	if err != nil {
		logger.ErrorContext(ctx, "复制文件内容失败", "error", err, "path", storedPath)
		// 尝试删除部分写入的文件
		_ = os.Remove(storedPath)
		return "", apperr.Wrap(err, apperr.CodeInternal, "无法写入文件内容")
	}

	logger.InfoContext(ctx, "文件成功保存到本地", "path", storedPath, "bytes_written", bytesWritten)
	return storedPath, nil
}

// DeleteFile 从本地磁盘删除指定路径的文件。
func (ls *LocalStorage) DeleteFile(ctx context.Context, storedPath string) error {
	// 安全性检查：确保路径在 uploadDir 下，防止删除任意文件
	absUploadDir, err := filepath.Abs(ls.uploadDir)
	if err != nil {
		logger.ErrorContext(ctx, "获取上传目录绝对路径失败", "error", err, "upload_dir", ls.uploadDir)
		return apperr.Wrap(err, apperr.CodeInternal, "无法验证文件路径")
	}
	absStoredPath, err := filepath.Abs(storedPath)
	if err != nil {
		logger.ErrorContext(ctx, "获取存储路径绝对路径失败", "error", err, "stored_path", storedPath)
		return apperr.Wrap(err, apperr.CodeInternal, "无法验证文件路径")
	}

	// Clean paths to handle potential ".." or "."
	cleanUploadDir := filepath.Clean(absUploadDir)
	cleanStoredPath := filepath.Clean(absStoredPath)

	if !filepath.HasPrefix(cleanStoredPath, cleanUploadDir) {
		logger.ErrorContext(ctx, "尝试删除上传目录之外的文件", "path", storedPath, "upload_dir", ls.uploadDir)
		return apperr.ErrPermissionDenied("无权删除指定路径的文件")
	}

	logger.InfoContext(ctx, "准备删除本地文件", "path", storedPath)
	err = os.Remove(storedPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.WarnContext(ctx, "尝试删除的文件不存在", "path", storedPath)
			return apperr.ErrNotFound("文件未找到，无法删除") // 或者返回 nil，认为删除不存在的文件是成功的
		}
		logger.ErrorContext(ctx, "删除本地文件失败", "error", err, "path", storedPath)
		return apperr.Wrap(err, apperr.CodeInternal, "无法删除文件")
	}

	logger.InfoContext(ctx, "本地文件删除成功", "path", storedPath)
	return nil
}

// GetFileReader 打开指定路径的文件并返回一个 io.ReadCloser。
func (ls *LocalStorage) GetFileReader(ctx context.Context, storedPath string) (io.ReadCloser, error) {
	// 同样进行路径安全检查
	absUploadDir, err := filepath.Abs(ls.uploadDir)
	if err != nil {
		logger.ErrorContext(ctx, "获取上传目录绝对路径失败", "error", err, "upload_dir", ls.uploadDir)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法验证文件路径")
	}
	absStoredPath, err := filepath.Abs(storedPath)
	if err != nil {
		logger.ErrorContext(ctx, "获取存储路径绝对路径失败", "error", err, "stored_path", storedPath)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法验证文件路径")
	}

	// Clean paths to handle potential ".." or "."
	cleanUploadDir := filepath.Clean(absUploadDir)
	cleanStoredPath := filepath.Clean(absStoredPath)

	if !filepath.HasPrefix(cleanStoredPath, cleanUploadDir) {
		logger.ErrorContext(ctx, "尝试读取上传目录之外的文件", "path", storedPath, "upload_dir", ls.uploadDir)
		return nil, apperr.ErrPermissionDenied("无权读取指定路径的文件")
	}

	logger.InfoContext(ctx, "准备打开本地文件读取", "path", storedPath)
	file, err := os.Open(storedPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.WarnContext(ctx, "尝试读取的文件不存在", "path", storedPath)
			return nil, apperr.ErrNotFound("文件未找到")
		}
		logger.ErrorContext(ctx, "打开本地文件失败", "error", err, "path", storedPath)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法打开文件")
	}

	// file 实现了 io.ReadCloser 接口
	return file, nil
}
