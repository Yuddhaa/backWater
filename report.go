package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReportData wraps the inputType to add summary statistics for the template
type ReportData struct {
	Title       string
	GeneratedAt string
	PassCount   int
	FailCount   int
	TotalCount  int
	SuccessRate int
	TotalTime   string // Added field for total execution time
	Data        inputType
}

// GenerateHTMLReport creates a beautiful HTML report from the test execution data.
// Updated signature to accept calculated stats and duration.
func GenerateHTMLReport(data inputType, totalTime string, total int, passed int) {
	// 1. Calculate derived statistics
	fail := total - passed
	rate := 0
	if total > 0 {
		rate = (passed * 100) / total
	}

	reportData := ReportData{
		Title:       data.Name,
		GeneratedAt: time.Now().Format("02-01-2006 15:04:05"),
		PassCount:   passed,
		FailCount:   fail,
		TotalCount:  total,
		SuccessRate: rate,
		TotalTime:   totalTime,
		Data:        data,
	}
	// printIndentJson("reportData", reportData)

	// 2. Define Template Functions (FuncMap)
	funcMap := template.FuncMap{
		"prettyJSON": func(v any) string {
			if v == nil {
				return ""
			}
			// If it's a string, it might be a JSON string, try to unmarshal it first
			if s, ok := v.(string); ok {
				if s == "" {
					return ""
				}
				var js map[string]any
				if err := json.Unmarshal([]byte(s), &js); err == nil {
					b, _ := json.MarshalIndent(js, "", "  ")
					return string(b)
				}
				return s
			}
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("%v", v)
			}
			return string(b)
		},
		"statusColor": func(actual, expected string) string {
			if strings.TrimSpace(actual) == strings.TrimSpace(expected) {
				return "bg-green-100 text-green-800 border-green-200"
			}
			return "bg-red-100 text-red-800 border-red-200"
		},
		"methodColor": func(method string) string {
			switch strings.ToUpper(method) {
			case "GET":
				return "bg-blue-100 text-blue-800"
			case "POST":
				return "bg-green-100 text-green-800"
			case "PUT":
				return "bg-yellow-100 text-yellow-800"
			case "DELETE":
				return "bg-red-100 text-red-800"
			default:
				return "bg-gray-100 text-gray-800"
			}
		},
		"isPass": func(actual, expected string) bool {
			return strings.TrimSpace(actual) == strings.TrimSpace(expected)
		},
	}

	// 3. Parse the Template
	// Note: You might want to pass the template path as an arg or keep it relative
	tmplContent, err := os.ReadFile(*template_file)
	if err != nil {
		fmt.Printf("Error reading template file: %v\n", err)
		return
	}

	t, err := template.New("report").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// 4. Create Output File
	if err := os.MkdirAll(*output_dir, 0o755); err != nil {
		log.Fatal(err)
	}
	timeStr := time.Now().Format("02-01_15.04")
	name := strings.ReplaceAll(data.Name, " ", "_")
	absPath, _ := filepath.Abs(*output_dir + "/" + name + "_" + timeStr + ".html")
	f, err := os.Create(absPath)
	if err != nil {
		fmt.Printf("Error creating report file: %v\n", err)
		return
	}
	defer f.Close()

	// 5. Execute
	if err := t.Execute(f, reportData); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
	}
	fmt.Printf("\nReport generated successfully at: %s\n", absPath)
}
