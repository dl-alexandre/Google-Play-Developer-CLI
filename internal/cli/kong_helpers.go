package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/xuri/excelize/v2"

	"github.com/dl-alexandre/gpd/internal/auth"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

const (
	formatJSON  = "json"
	formatTable = "table"
	formatExcel = "excel"
)

// newAuthManager creates a new auth manager instance.
func newAuthManager() *auth.Manager {
	secureStorage := storage.New()
	return auth.NewManager(secureStorage)
}

// outputResult formats and outputs a result based on the format.
func outputResult(result *output.Result, format string, pretty bool) error {
	switch format {
	case formatJSON:
		return outputJSON(result, pretty)
	case formatTable:
		return outputTable(result)
	case formatExcel:
		return outputExcel(result)
	default:
		return outputJSON(result, pretty)
	}
}

// outputJSON outputs result as JSON.
func outputJSON(result *output.Result, pretty bool) error {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(result, "", "  ")
	} else {
		data, err = json.Marshal(result)
	}

	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

// outputTable outputs result as a table.
func outputTable(result *output.Result) error {
	// Extract data from result
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		// Fall back to JSON if we can't format as table
		return outputJSON(result, false)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Key", "Value"})

	for key, value := range data {
		_ = table.Append([]string{key, fmt.Sprintf("%v", value)})
	}

	_ = table.Render()
	return nil
}

// outputExcel outputs result as Excel (.xlsx) format.
func outputExcel(result *output.Result) error {
	if result.Error != nil {
		return outputJSON(result, false)
	}

	data := result.Data
	if data == nil {
		return outputJSON(result, false)
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data"
	f.SetSheetName("Sheet1", sheetName)

	switch v := data.(type) {
	case []interface{}:
		if err := writeExcelSlice(f, sheetName, v); err != nil {
			return err
		}
	case []map[string]interface{}:
		slice := make([]interface{}, len(v))
		for i, item := range v {
			slice[i] = item
		}
		if err := writeExcelSlice(f, sheetName, slice); err != nil {
			return err
		}
	case map[string]interface{}:
		if err := writeExcelMap(f, sheetName, v); err != nil {
			return err
		}
	default:
		return outputJSON(result, false)
	}

	if result.Meta != nil && (len(result.Meta.Warnings) > 0 || result.Meta.DurationMs > 0) {
		if err := writeExcelMetadata(f, result.Meta); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return err
	}

	_, err := os.Stdout.Write(buf.Bytes())
	return err
}

func writeExcelSlice(f *excelize.File, sheetName string, data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot convert data to Excel format")
	}

	var headers []string
	for k := range first {
		headers = append(headers, k)
	}

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, boldStyle)
	}

	for rowIdx, item := range data {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		for colIdx, header := range headers {
			cell := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), rowIdx+2)
			value := row[header]
			f.SetCellValue(sheetName, cell, value)
		}
	}

	for i := range headers {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	return nil
}

func writeExcelMap(f *excelize.File, sheetName string, data map[string]interface{}) error {
	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	rowIdx := 1
	for key, value := range data {
		keyCell := fmt.Sprintf("A%d", rowIdx)
		valueCell := fmt.Sprintf("B%d", rowIdx)

		f.SetCellValue(sheetName, keyCell, key)
		f.SetCellStyle(sheetName, keyCell, keyCell, boldStyle)
		f.SetCellValue(sheetName, valueCell, value)

		rowIdx++
	}

	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 30)

	return nil
}

func writeExcelMetadata(f *excelize.File, meta *output.Metadata) error {
	sheetName := "Metadata"
	f.NewSheet(sheetName)

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	rowIdx := 1

	if meta.DurationMs > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), "Duration (ms)")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), boldStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), meta.DurationMs)
		rowIdx++
	}

	if len(meta.Services) > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), "Services")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), boldStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("%v", meta.Services))
		rowIdx++
	}

	if meta.NoOp {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), "NoOp")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), boldStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), "true")
		if meta.NoOpReason != "" {
			rowIdx++
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), "NoOp Reason")
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), boldStyle)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), meta.NoOpReason)
		}
		rowIdx++
	}

	if len(meta.Warnings) > 0 {
		rowIdx += 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), "Warnings")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), boldStyle)
		rowIdx++
		for _, warning := range meta.Warnings {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), warning)
			rowIdx++
		}
	}

	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 40)

	return nil
}

// requirePackage validates that a package name is provided.
func requirePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name is required")
	}
	return nil
}
