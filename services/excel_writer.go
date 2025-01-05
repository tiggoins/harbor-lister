package services

import (
	"fmt"
	"sync"

	"github.com/tiggoins/harbor-lister/types"
	"github.com/xuri/excelize/v2"
)

type ExcelWriter struct {
	mu     sync.Mutex
	file   *excelize.File
	sheet  string
	rowNum int
}

func NewExcelWriter() *ExcelWriter {
	f := excelize.NewFile()
	sheet := "Harbor镜像列表"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// 设置表头
	f.SetCellValue(sheet, "A1", "项目")
	f.SetCellValue(sheet, "B1", "仓库")
	f.SetCellValue(sheet, "C1", "标签")
	f.SetCellValue(sheet, "D1", "推送时间")

	// 设置列宽
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 50)
	f.SetColWidth(sheet, "D", "D", 20)

	return &ExcelWriter{
		file:   f,
		sheet:  sheet,
		rowNum: 2, // 从第2行开始写数据
	}
}

func (w *ExcelWriter) WriteProject(project string, repos types.ProjectMap) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	startRow := w.rowNum

	for repo, tags := range repos {
		// 计算当前仓库的行数
		repoStartRow := w.rowNum
		tagsCount := len(tags)

		// 写入标签和推送时间
		for _, tag := range tags {
			w.file.SetCellValue(w.sheet, fmt.Sprintf("C%d", w.rowNum), tag[0].Name)
			w.file.SetCellValue(w.sheet, fmt.Sprintf("D%d", w.rowNum), tag[0].PushTime)
			w.rowNum++
		}

		// 如果有多个标签，合并仓库单元格
		if tagsCount > 1 {
			w.file.MergeCell(w.sheet, fmt.Sprintf("B%d", repoStartRow),
				fmt.Sprintf("B%d", w.rowNum-1))
		}
		w.file.SetCellValue(w.sheet, fmt.Sprintf("B%d", repoStartRow), repo)
	}

	// 合并项目单元格
	if w.rowNum > startRow {
		w.file.MergeCell(w.sheet, fmt.Sprintf("A%d", startRow),
			fmt.Sprintf("A%d", w.rowNum-1))
		w.file.SetCellValue(w.sheet, fmt.Sprintf("A%d", startRow), project)
	}

	return nil
}

func (w *ExcelWriter) Save(filename string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.SaveAs(filename)
}
