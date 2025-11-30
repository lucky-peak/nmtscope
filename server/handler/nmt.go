package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/lucky-peak/nmtscope/server/config"
	"github.com/lucky-peak/nmtscope/server/parser"
	"github.com/lucky-peak/nmtscope/server/types"
)

func NMTHandler(w http.ResponseWriter, r *http.Request) {
	// log.Printf("Received request for NMT reports: %v", r.URL.Query()) // 如果需要打印，使用 log

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	begin := r.URL.Query().Get("begin")
	end := r.URL.Query().Get("end")

	if begin == "" || end == "" {
		http.Error(w, errors.New("begin and end are required").Error(), http.StatusBadRequest)
		return
	}

	beginTS, err := strconv.ParseInt(begin, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("invalid begin timestamp: %w", err).Error(), http.StatusBadRequest)
		return
	}
	endTs, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("invalid end timestamp: %w", err).Error(), http.StatusBadRequest)
		return
	}

	if beginTS >= endTs {
		http.Error(w, errors.New("begin must be less than end").Error(), http.StatusBadRequest)
		return
	}

	// ListNMTReports 应该被认为是一个可能出错的操作。
	// 虽然在这个示例中没有返回错误，但在实际项目中，文件操作应返回错误。
	// 假设 ListNMTReports 内部已处理错误或返回一个空切片。
	nmtReports := parser.ListNMTReports(config.CONFIG.ReportDir, beginTS, endTs)

	res := types.NMTResponse[[]types.NMTReport]{
		Data: nmtReports,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
