package utils

import (
	"math"
	"mbook/conf"

	//"so-translate/conf"
)

type Pagination struct {
	CurrentPageNum  int    `json:"currentPageNum"`
	PageSize        int    `json:"pageSize"`
	PageCount       int    `json:"pageCount"`
	WindowSize      int    `json:"windowSize"`
	RecordCount     int    `json:"recordCount"`
	PageNums        []int  `json:"pageNums"`
	NextPageNum     int    `json:"nextPageNum"`
	PreviousPageNum int    `json:"previousPageNum"`
	FirstPageNum    int    `josn:"firstPageNum"`
	LastPageNum     int    `json:"lastPageNum"`
	PageURL         string `json:"pageURL"`
}

// NewPagination creates a new pagination with the specified current page num and record count.
func NewPagination(currentPageNum, recordCount int) *Pagination {
	PageSize := conf.OSPageSize
	pageCount := int(math.Ceil(float64(recordCount) / float64(PageSize)))

	previousPageNum := currentPageNum - 1
	if 1 > previousPageNum {
		previousPageNum = 0
	}
	nextPageNum := currentPageNum + 1
	if nextPageNum > pageCount {
		nextPageNum = 0
	}
	WindowSize := conf.WindowSize
	pageNums := paginate(currentPageNum, pageCount, WindowSize)
	firstPageNum := 0
	lastPageNum := 0
	if 0 < len(pageNums) {
		firstPageNum = pageNums[0]
		lastPageNum = pageNums[len(pageNums)-1]
	}

	return &Pagination{
		CurrentPageNum:  currentPageNum,
		NextPageNum:     nextPageNum,
		PreviousPageNum: previousPageNum,
		PageSize:        PageSize,
		PageCount:       pageCount,
		WindowSize:      WindowSize,
		RecordCount:     recordCount,
		PageNums:        pageNums,
		FirstPageNum:    firstPageNum,
		LastPageNum:     lastPageNum,
	}
}

func paginate(currentPageNum, pageCount, windowSize int) []int {
	var ret []int

	if pageCount < windowSize {
		for i := 0; i < pageCount; i++ {
			ret = append(ret, i+1)
		}
	} else {
		first := currentPageNum + 1 - windowSize/2
		if first < 1 {
			first = 1
		}
		if first+windowSize > pageCount {
			first = pageCount - windowSize + 1
		}
		for i := 0; i < windowSize; i++ {
			ret = append(ret, first+i)
		}
	}

	return ret
}
