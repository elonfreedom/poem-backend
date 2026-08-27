package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigrationsSynced 确保嵌入的迁移文件与根目录 migrations/ 保持同步
// 防止 pkg/migrate/migrations/ 和根目录 migrations/ 不一致导致迁移失败
func TestMigrationsSynced(t *testing.T) {
	// 读取嵌入的迁移文件
	embeddedFiles := make(map[string]string)
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("读取嵌入迁移文件失败: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			t.Fatalf("读取嵌入文件 %s 失败: %v", entry.Name(), err)
		}
		embeddedFiles[entry.Name()] = string(data)
	}

	// 读取根目录的迁移文件
	rootDir := filepath.Join("..", "..", "migrations")
	rootFiles := make(map[string]string)
	entries, err = os.ReadDir(rootDir)
	if err != nil {
		t.Fatalf("读取根目录迁移文件失败: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(rootDir, entry.Name()))
		if err != nil {
			t.Fatalf("读取根目录文件 %s 失败: %v", entry.Name(), err)
		}
		rootFiles[entry.Name()] = string(data)
	}

	// 检查数量是否一致
	if len(embeddedFiles) != len(rootFiles) {
		t.Errorf("迁移文件数量不一致: 嵌入 %d 个, 根目录 %d 个", len(embeddedFiles), len(rootFiles))
	}

	// 检查每个文件内容是否一致
	for name, content := range rootFiles {
		embedded, ok := embeddedFiles[name]
		if !ok {
			t.Errorf("嵌入文件中缺少 %s", name)
			continue
		}
		if embedded != content {
			t.Errorf("文件 %s 内容不一致，请运行: cp migrations/*.sql pkg/migrate/migrations/", name)
		}
	}
}
