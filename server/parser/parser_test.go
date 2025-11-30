package parser

import (
	"bufio"
	"os"
	"testing"

	types "github.com/lucky-peak/nmtscope/server/types"
)

func Test_parseNMTEntries(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		lines []string
		want  []types.NMTEntry
	}{
		// TODO: Add test cases.
		{
			name: "Total header",
			// [{{Total 5813841 382817}} {{Java Heap 4194304 157696}} {{Class 1049451 5099}} {{Thread 88189 88189}} {{Code 252156 18684}} {{GC 132525 53741}} {{GCCardSet 33 33}} {{Compiler 193 193}} {{Internal 352 352}} {{Other 18 18}} {{Symbol 10576 10576}} {{Native Memory Tracking 2879 2879}} {{Shared class space 16384 13904}} {{Arena Chunk 2 2}} {{Module 206 206}} {{Safepoint 32 32}} {{Synchronization 822 822}} {{Serviceability 18 18}} {{Metaspace 65699 30371}} {{String Deduplication 1 1}} {{Object Monitors 0 0}} {{Unknown 0 0}}], want [{{Total 1234567890 1234567890}}]
			want: []types.NMTEntry{
				{
					NMTHeader: types.NMTHeader{
						Name:      "Total",
						Reserved:  5813841,
						Committed: 382817,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Java Heap",
						Reserved:  4194304,
						Committed: 157696,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Class",
						Reserved:  1049451,
						Committed: 5099,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Thread",
						Reserved:  88189,
						Committed: 88189,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Code",
						Reserved:  252156,
						Committed: 18684,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "GC",
						Reserved:  132525,
						Committed: 53741,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "GCCardSet",
						Reserved:  33,
						Committed: 33,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Compiler",
						Reserved:  193,
						Committed: 193,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Internal",
						Reserved:  352,
						Committed: 352,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Other",
						Reserved:  18,
						Committed: 18,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Symbol",
						Reserved:  10576,
						Committed: 10576,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Native Memory Tracking",
						Reserved:  2879,
						Committed: 2879,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Shared class space",
						Reserved:  16384,
						Committed: 13904,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Arena Chunk",
						Reserved:  2,
						Committed: 2,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Module",
						Reserved:  206,
						Committed: 206,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Safepoint",
						Reserved:  32,
						Committed: 32,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Synchronization",
						Reserved:  822,
						Committed: 822,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Serviceability",
						Reserved:  18,
						Committed: 18,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Metaspace",
						Reserved:  65699,
						Committed: 30371,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "String Deduplication",
						Reserved:  1,
						Committed: 1,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Object Monitors",
						Reserved:  0,
						Committed: 0,
					},
				},
				{
					NMTHeader: types.NMTHeader{
						Name:      "Unknown",
						Reserved:  0,
						Committed: 0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open("../fixtures/nmt_53732_1764484822.txt")
			if err != nil {
				t.Errorf("os.Open() error = %v", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			got := parseNMTEntries(lines)
			if len(got) != len(tt.want) {
				t.Errorf("parseNMTEntries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFileName(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fileName string
		want     int
		want2    int64
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "Valid file name",
			fileName: "nmt_53732_1764484822.txt",
			want:     53732,
			want2:    1764484822,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2, gotErr := parseReportFileName(tt.fileName)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseFileName() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseFileName() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("parseFileName() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("parseFileName() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_parseNMTReportMetadata(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reportFilepath string
		want           types.NMTReport
		wantErr        bool
	}{
		// TODO: Add test cases.
		{
			name:           "Valid report file path",
			reportFilepath: "../fixtures/nmt_53732_1764484822.txt",
			want: types.NMTReport{
				PID:     53732,
				Created: 1764484822,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseNMTReportMetadata(tt.reportFilepath)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseNMTReport() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseNMTReport() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got.PID != tt.want.PID || got.Created != tt.want.Created {
				t.Errorf("parseNMTReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListNMTReports(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		dir   string
		begin int
		end   int
		want  []types.NMTReport
	}{
		// TODO: Add test cases.
		{
			name:  "Valid directory",
			dir:   "../fixtures",
			begin: 0,
			end:   1764484823,
			want: []types.NMTReport{
				{
					PID:     53732,
					Created: 1764484822,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ListNMTReports(tt.dir, int64(tt.begin), int64(tt.end))
			// TODO: update the condition below to compare got with tt.want.
			if len(got) != len(tt.want) {
				t.Errorf("ListNMTReports() = %v, want %v", got, tt.want)
			}
		})
	}
}
