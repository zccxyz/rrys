package model

import "time"

type MoviesDownload struct {
	Vid       uint64
	Season    uint64 //第几季
	Episode   uint64 //第几集
	Yyets     string //人人影视
	Dl        string //电驴
	Cl        string //磁力
	Ctwp      string //城通网盘
	Wy        string //微云
	Size      string
	ClPsd     string
	DlPsd     string
	CtwpPsd   string
	WyPsd     string
	FileName  string
	Baidu     string
	BaiduPsd  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
