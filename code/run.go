package code

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"rrys/model"
	"strconv"
	"sync"
	"time"
)

const (
	host  = "http://pc.zmzapi.com/index.php?g=api/pv3&m=index&client=5&accesskey=519f9cab85c8059d17544947k361a827&"
	limit = 30
	order = "itemupdate"
)

var (
	db *gorm.DB
)

func init() {
	db = model.GetDb()
}

//爬取视频
func VideoRun() {
	tvStatus := false //tv是否完成
	mvStatus := false //mv是否完成
	channel := "tv"
	page := 1
	for {
		if tvStatus && mvStatus {
			return
		}
		//从第一页开始获取数据
		rs, err := GetData(page, channel)
		if err != nil {
			log.Fatal(err)
		}
		jsonData := gjson.Parse(string(rs))
		if jsonData.Get("status").Int() != 1 {
			log.Fatal("请求出错", page, "tv")
		}

		//新数据获取详情页的数据并保存数据库
		var wg sync.WaitGroup
		for _, v := range jsonData.Get("data.list").Array() {
			wg.Add(1)
			vid := v.Get("id").Uint()
			//判断视频是否存在
			hasMv := model.Movies{}
			if !db.Where("vid = ?", vid).First(&hasMv).RecordNotFound() {
				wg.Done()
				continue
			}

			detailRs, err := GetDetail(vid)
			if err != nil {
				log.Println(err)
				wg.Done()
			}
			allJson := gjson.Parse(string(detailRs))
			if allJson.Get("status").Int() != 1 {
				echo("详情请求出错-" + v.Get("id").String() + "-" + allJson.Get("info").String())
				wg.Done()
				continue
			}
			detailRsJson := allJson.Get("data.detail")
			mv := model.Movies{
				Vid:            detailRsJson.Get("id").Uint(),
				Cnname:         detailRsJson.Get("cnname").String(),
				Enname:         detailRsJson.Get("enname").String(),
				Channel:        detailRsJson.Get("channel").String(),
				Area:           detailRsJson.Get("area").String(),
				Category:       detailRsJson.Get("category").String(),
				Tvstation:      detailRsJson.Get("tvstation").String(),
				Lang:           detailRsJson.Get("lang").String(),
				PlayStatus:     detailRsJson.Get("play_status").String(),
				Rank:           detailRsJson.Get("rank").Uint(),
				Views:          v.Get("views").Uint(),
				Score:          detailRsJson.Get("score").Float(),
				PublishYear:    v.Get("publish_year").Uint(),
				Itemupdate:     v.Get("itemupdate").Uint(),
				Poster:         detailRsJson.Get("poster").String(),
				FavoriteStatus: detailRsJson.Get("favorite_status").Uint(),
				Season:         v.Get("last_episode.season").Uint(),
				Episode:        v.Get("last_episode.episode").Uint(),
				Premiere:       detailRsJson.Get("premiere").String(),
				Zimuzu:         detailRsJson.Get("zimuzu").String(),
				Aliasname:      detailRsJson.Get("aliasname").String(),
				ScoreCounts:    detailRsJson.Get("score_counts").Uint(),
				Content:        detailRsJson.Get("content").String(),
				CloseResource:  detailRsJson.Get("close_resource").Uint(),
				Website:        detailRsJson.Get("website").String(),
				Level:          detailRsJson.Get("level").String(),
				Director:       detailRsJson.Get("director").String(),
				Writer:         detailRsJson.Get("writer").String(),
				Actor:          detailRsJson.Get("actor").String(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if dbErr := db.Create(&mv).Error; dbErr != nil {
				echo(dbErr.Error() + "-" + detailRsJson.Get("id").String())
				wg.Done()
				continue
			}
			//保存下载地址
			listJson := allJson.Get("data.list")
			listArr := listJson.Array()
			if len(listArr) == 0 {
				wg.Done()
				continue
			}

			for _, v := range listArr {
				episodes := v.Get("episodes").Array()
				if len(episodes) == 0 {
					continue
				}

				for _, v2 := range episodes {
					md := model.MoviesDownload{
						Vid:       mv.Vid,
						Season:    v.Get("season").Uint(),
						Episode:   v2.Get("episode").Uint(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					if v2.Get("files.APP").Exists() {
						md.Wy = v2.Get("files.APP.address").String()
						md.WyPsd = v2.Get("files.APP.passwd").String()
					}
					if v2.Get("files.yyets").Exists() {
						md.FileName = v2.Get("files.yyets.file_name").String()
						md.Yyets = v2.Get("files.yyets.name").String()
					}
					mp4 := v2.Get("files.MP4").Exists()
					hr := v2.Get("files.HR-HDTV").Exists()
					rmvb := v2.Get("files.RMVB").Exists()
					if mp4 || hr || rmvb {
						c := "HR-HDTV"
						if !hr {
							if !mp4 {
								c = "RMVB"
							} else {
								c = "MP4"
							}
						}
						for _, v3 := range v2.Get("files." + c).Array() {
							if md.FileName == "" {
								md.FileName = v3.Get("name").String()
							}
							md.Size = v3.Get("size").String()
							if v3.Get("way").Uint() == 1 {
								//电驴
								md.Dl = v3.Get("address").String()
								md.DlPsd = v3.Get("passwd").String()
							} else if v3.Get("way").Uint() == 12 {
								md.Ctwp = v3.Get("address").String()
								md.CtwpPsd = v3.Get("passwd").String()
							} else if v3.Get("way").Uint() == 2 {
								md.Wy = v3.Get("address").String()
								md.WyPsd = v3.Get("passwd").String()
							} else if v3.Get("way").Uint() == 9 {
								md.Baidu = v3.Get("address").String()
								md.BaiduPsd = v3.Get("passwd").String()
							}
						}
					}
					if dbErr := db.Create(&md).Error; dbErr != nil {
						echo(dbErr.Error() + "下载地址保存失败")
						continue
					}
				}
			}

			wg.Done()
		}
		wg.Wait()
		echo("【" + channel + "】第" + strconv.Itoa(page) + "页~~~~~~~~~~~~~~~~~~~~~~")
		page++
		//判断是否最后一页
		if len(jsonData.Get("data.list").Array()) < limit {
			if tvStatus {
				mvStatus = true
			} else {
				channel = "movie"
				page = 1
				tvStatus = true
			}
		}
	}
}

//更新视频
func UpdateVideo() {
	for {
		time.Sleep(time.Minute)
		//每天23点更新视频
		now := time.Now()
		if now.Hour() == 23 {
			//获取更新视频数据
			rs, err := getUpdateVideo()
			if err != nil {
				log.Println("获取更新视频数据失败: ", err)
				continue
			}
			updateJson := gjson.Parse(string(rs))
			status := updateJson.Get("status").Int()
			if status != 1 {
				log.Printf("status: %d, info: %s \n", status, updateJson.Get("info").String())
				continue
			}
			list := updateJson.Get("data").Array()
			//获取今日更新视频
			var need gjson.Result
			ok := false
			for _, v := range list {
				if ok {
					break
				}
				for _, v2 := range v.Array() {
					if v2.Get("date").String() == now.Format("2006-01-02") {
						need = v2
						ok = true
						break
					}
				}
			}
			var wg sync.WaitGroup
			for _, v := range need.Get("list").Array() {
				wg.Add(1)
				fmt.Println(v.Get("cnname").String())
				go saveVideo(v.Get("id").Uint(), &wg, v.Get("season").Uint(), v.Get("episode").Uint())
			}
			wg.Wait()
			echo(now.Format("2006-01-02") + "更新完毕")
			time.Sleep(time.Minute * 30)
		}
	}
}

//保存一部电视剧
func saveVideo(vid uint64, wg *sync.WaitGroup, season uint64, episode uint64) {
	hasMv := model.Movies{}

	detailRs, err := GetDetail(vid)
	if err != nil {
		log.Println(err)
		wg.Done()
	}
	allJson := gjson.Parse(string(detailRs))
	if allJson.Get("status").Int() != 1 {
		echo("详情请求出错-" + string(vid) + "-" + allJson.Get("info").String())
		wg.Done()
	}
	detailRsJson := allJson.Get("data.detail")
	//判断视频是否存在
	if db.Where("vid = ?", vid).First(&hasMv).RecordNotFound() {
		//电影不存在则保存
		mv := model.Movies{
			Vid:            detailRsJson.Get("id").Uint(),
			Cnname:         detailRsJson.Get("cnname").String(),
			Enname:         detailRsJson.Get("enname").String(),
			Channel:        detailRsJson.Get("channel").String(),
			Area:           detailRsJson.Get("area").String(),
			Category:       detailRsJson.Get("category").String(),
			Tvstation:      detailRsJson.Get("tvstation").String(),
			Lang:           detailRsJson.Get("lang").String(),
			PlayStatus:     detailRsJson.Get("play_status").String(),
			Rank:           detailRsJson.Get("rank").Uint(),
			Views:          0,
			Score:          detailRsJson.Get("score").Float(),
			PublishYear:    0,
			Itemupdate:     0,
			Poster:         detailRsJson.Get("poster").String(),
			FavoriteStatus: detailRsJson.Get("favorite_status").Uint(),
			Season:         season,
			Episode:        episode,
			Premiere:       detailRsJson.Get("premiere").String(),
			Zimuzu:         detailRsJson.Get("zimuzu").String(),
			Aliasname:      detailRsJson.Get("aliasname").String(),
			ScoreCounts:    detailRsJson.Get("score_counts").Uint(),
			Content:        detailRsJson.Get("content").String(),
			CloseResource:  detailRsJson.Get("close_resource").Uint(),
			Website:        detailRsJson.Get("website").String(),
			Level:          detailRsJson.Get("level").String(),
			Director:       detailRsJson.Get("director").String(),
			Writer:         detailRsJson.Get("writer").String(),
			Actor:          detailRsJson.Get("actor").String(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if dbErr := db.Create(&mv).Error; dbErr != nil {
			echo(dbErr.Error() + "-" + detailRsJson.Get("id").String())
			wg.Done()
		}
	} else {
		if dbErr := db.Table("movies").Where("vid = ?", vid).
			Updates(model.Movies{Season: season, Episode: episode}).Error; dbErr != nil {
			echo(dbErr.Error() + "-" + detailRsJson.Get("id").String())
			wg.Done()
		}
	}
	//保存下载地址
	listJson := allJson.Get("data.list")
	listArr := listJson.Array()
	if len(listArr) == 0 {
		wg.Done()
	}
	for _, v := range listArr {
		episodes := v.Get("episodes").Array()
		if len(episodes) == 0 {
			continue
		}

		for _, v2 := range episodes {
			md := model.MoviesDownload{
				Vid:       vid,
				Season:    v.Get("season").Uint(),
				Episode:   v2.Get("episode").Uint(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if v2.Get("files.APP").Exists() {
				md.Wy = v2.Get("files.APP.address").String()
				md.WyPsd = v2.Get("files.APP.passwd").String()
			}
			if v2.Get("files.yyets").Exists() {
				md.FileName = v2.Get("files.yyets.file_name").String()
				md.Yyets = v2.Get("files.yyets.name").String()
			}
			mp4 := v2.Get("files.MP4").Exists()
			hr := v2.Get("files.HR-HDTV").Exists()
			rmvb := v2.Get("files.RMVB").Exists()
			if mp4 || hr || rmvb {
				c := "HR-HDTV"
				if !hr {
					if !mp4 {
						c = "RMVB"
					} else {
						c = "MP4"
					}
				}
				for _, v3 := range v2.Get("files." + c).Array() {
					if md.FileName == "" {
						md.FileName = v3.Get("name").String()
					}
					md.Size = v3.Get("size").String()
					if v3.Get("way").Uint() == 1 {
						//电驴
						md.Dl = v3.Get("address").String()
						md.DlPsd = v3.Get("passwd").String()
					} else if v3.Get("way").Uint() == 12 {
						md.Ctwp = v3.Get("address").String()
						md.CtwpPsd = v3.Get("passwd").String()
					} else if v3.Get("way").Uint() == 2 {
						md.Wy = v3.Get("address").String()
						md.WyPsd = v3.Get("passwd").String()
					} else if v3.Get("way").Uint() == 9 {
						md.Baidu = v3.Get("address").String()
						md.BaiduPsd = v3.Get("passwd").String()
					}
				}
			}
			hasMd := model.MoviesDownload{}
			if db.Where("vid = ? and season = ? and episode = ?", vid, md.Season, md.Episode).
				First(&hasMd).RecordNotFound() {
				if dbErr := db.Create(&md).Error; dbErr != nil {
					echo(dbErr.Error() + "下载地址保存失败")
				}
			}
		}
	}
	wg.Done()
}

func GetData(page int, channel string) ([]byte, error) {
	url := host + "order=" + order + "&a=resource_storage&limit=" + strconv.Itoa(limit) + "&page=" + strconv.Itoa(page) + "&channel=" + channel
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func GetDetail(vid uint64) ([]byte, error) {
	url := host + "a=resource&id=" + strconv.FormatUint(vid, 10)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

//视频更新数据
func getUpdateVideo() ([]byte, error) {
	url := host + "a=episode_list"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func echo(str string) {
	_, _ = os.Stdout.WriteString(str + "\n")
}
