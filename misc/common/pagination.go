package common

//用于分页计算
type Pagination struct{
	pageNo int32
	pageNum int32
	totalCount int32
	maxPageNo int32
}

func NewPagination(totalCount, pageNum int32) *Pagination {
	var ob = Pagination{
		pageNum:    pageNum,
		totalCount: totalCount,
		pageNo: 1,
	}
	ob.maxPageNo = (totalCount + pageNum - 1) / pageNum
	return &ob
}

func (this *Pagination) SetPageNo(pageNo int32) {
	this.pageNo = pageNo
	if this.maxPageNo < pageNo {
		this.pageNo = this.maxPageNo
	} else if pageNo < 1 {
		this.pageNo = 1
	}
}

func (this *Pagination) GetPageNo() int32 {
	return this.pageNo
}

func (this *Pagination) GetCurrentPageIndex() (beginIndex, endIndex int32) {
	beginIndex = (this.pageNo - 1) * this.pageNum
	endIndex = this.pageNo * this.pageNum
	if endIndex > this.totalCount{
		endIndex = this.totalCount
	}
	return
}

