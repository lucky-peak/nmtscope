package parser

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucky-peak/nmtscope/server/config"
	constants "github.com/lucky-peak/nmtscope/server/constants"
	types "github.com/lucky-peak/nmtscope/server/types"
	"github.com/lucky-peak/nmtscope/server/utils"
)

func ListNMTReports(dir string, beginTS int64, endTS int64) (nmtReports []types.NMTReport) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".txt") {
			return nil
		}
		nmtReportMetadata, err := parseNMTReportMetadata(path)
		if err != nil {
			return err
		}

		if nmtReportMetadata.Created < time.Now().Add(-time.Duration(config.CONFIG.Retention)*time.Minute).Unix() {
			// 异步删除过期文件
			go func() {
				if err = os.Remove(path); err != nil {
					log.Printf("Error deleting file %s: %v", path, err)
				}
			}()
			return nil
		}

		if nmtReportMetadata.Created < beginTS || nmtReportMetadata.Created > endTS {
			return nil
		}

		lines, err := utils.ReadAll(path)
		if err != nil {
			return err
		}
		nmtReportMetadata.NMTEntries = parseNMTEntries(lines)
		nmtReports = append(nmtReports, nmtReportMetadata)
		return nil
	})
	return nmtReports
}

func parseNMTReportMetadata(reportFilepath string) (nmtReport types.NMTReport, err error) {
	reportFilename := filepath.Base(reportFilepath)

	pid, ts, err := parseReportFileName(reportFilename)
	if err != nil {
		return
	}
	nmtReport.PID = pid
	nmtReport.Created = ts
	return nmtReport, nil
}

// nmt_53732_1764484822.txt
func parseReportFileName(reportFileName string) (pid int, ts int64, err error) {
	reportFileName = utils.TrimSpace(reportFileName)
	reportFileName, _ = strings.CutSuffix(reportFileName, ".txt")

	parts := strings.Split(reportFileName, "_")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("invalid file name format: %s", reportFileName)
	}

	if parts[0] != "nmt" {
		return 0, 0, fmt.Errorf("invalid file name format: %s", reportFileName)
	}

	pid = utils.ParseInt(parts[1])
	if pid <= 0 {
		return 0, 0, fmt.Errorf("invalid pid: %s", parts[1])
	}

	ts = utils.ParseInt64(parts[2])
	if ts <= 0 {
		return 0, 0, fmt.Errorf("invalid ts: %s", parts[2])
	}

	return pid, ts, nil
}

func parseNMTEntries(lines []string) []types.NMTEntry {
	var nmtEntries []types.NMTEntry
	for _, line := range lines {
		line = utils.TrimSpace(line)
		totalHeaderMatch := constants.NMT_TOTAL_HEADER_REGEX.FindAllStringSubmatch(line, -1)
		if len(totalHeaderMatch) > 0 {
			nmtEntries = append(nmtEntries, types.NMTEntry{
				NMTHeader: types.NMTHeader{
					Name:      "Total",
					Reserved:  utils.ParseInt(totalHeaderMatch[0][1]),
					Committed: utils.ParseInt(totalHeaderMatch[0][2]),
				},
			})
			continue
		}

		categoryMatch := constants.NMT_CATEGORY_HEADER_REGEX.FindAllStringSubmatch(line, -1)
		if len(categoryMatch) > 0 {
			nmtEntries = append(nmtEntries, types.NMTEntry{
				NMTHeader: types.NMTHeader{
					Name:      categoryMatch[0][1],
					Reserved:  utils.ParseInt(categoryMatch[0][2]),
					Committed: utils.ParseInt(categoryMatch[0][3]),
				},
			})
			continue
		}
	}
	return nmtEntries
}
