package services

import (
	"fmt"
	"sync"

	"github.com/tiggoins/harbor-lister/types"
	"github.com/tiggoins/harbor-lister/utils"
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

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E90FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "D1", headerStyle)

	// 锁定表头行
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
	})

	return &ExcelWriter{
		file:   f,
		sheet:  sheet,
		rowNum: 2,
	}
}

func (w *ExcelWriter) WriteProject(project *types.Project) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 设置单元格样式
	cellStyle, _ := w.file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	startRow := w.rowNum

	// 遍历仓库列表
	for _, repo := range project.Repositories {
		repoStartRow := w.rowNum
		artifactCount := len(repo.Artifact)

		// 遍历每个 Artifact 的标签
		for _, artifact := range repo.Artifact {
			for _, tag := range artifact.Tags {
				cellRow := fmt.Sprintf("%d", w.rowNum)
				w.file.SetCellValue(w.sheet, fmt.Sprintf("C%s", cellRow), tag.Name)
				w.file.SetCellValue(w.sheet, fmt.Sprintf("D%s", cellRow),
					utils.FormatTime(tag.PushTime))

				// 设置样式
				w.file.SetCellStyle(w.sheet, fmt.Sprintf("C%s", cellRow), fmt.Sprintf("D%s", cellRow), cellStyle)
				w.rowNum++
			}
		}

		// 如果有多个 Artifact 或标签，合并仓库单元格
		if artifactCount > 1 || w.rowNum-repoStartRow > 1 {
			err := w.file.MergeCell(w.sheet,
				fmt.Sprintf("B%d", repoStartRow),
				fmt.Sprintf("B%d", w.rowNum-1))
			if err != nil {
				return fmt.Errorf("合并仓库单元格失败: %v", err)
			}
		}
		w.file.SetCellValue(w.sheet, fmt.Sprintf("B%d", repoStartRow), repo.Name)
		w.file.SetCellStyle(w.sheet, fmt.Sprintf("B%d", repoStartRow), fmt.Sprintf("B%d", repoStartRow), cellStyle)
	}

	// 如果当前项目有数据，合并项目单元格
	if w.rowNum > startRow {
		err := w.file.MergeCell(w.sheet,
			fmt.Sprintf("A%d", startRow),
			fmt.Sprintf("A%d", w.rowNum-1))
		if err != nil {
			return fmt.Errorf("合并项目单元格失败: %v", err)
		}
		w.file.SetCellValue(w.sheet, fmt.Sprintf("A%d", startRow), project.Name)
		w.file.SetCellStyle(w.sheet, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", startRow), cellStyle)
	}

	return nil
}

func (w *ExcelWriter) Save(filename string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.SaveAs(filename)
}
