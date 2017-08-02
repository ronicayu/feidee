package feidee

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//账单科目
type Category struct {
	IdName
	Type  int  //科目类别：支出或收入，参见TranTypeXXX常量
	IsSub bool //是否是子科目
}

//初始化账本、分类、账户、商家、项目、成员等信息
func (cli *Client) SyncMetaInfo() error {
	resp, err := cli.httpClient.Get(BaseUrl + "/money/tally/new.do")
	if err != nil {
		return fmt.Errorf("请求出错: %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return fmt.Errorf("读取响应出错: %s", err)
	}

	div := doc.Find("div#filter-bar div.fb-choose")

	//解析科目信息
	cli.CategoryMap = parseCategoryMap(div)

	//解析商家、成员、账户、项目信息
	cli.StoreMap = parseIdNameMap(div, "store")
	cli.MemberMap = parseIdNameMap(div, "member")
	cli.AccountMap = parseIdNameMap(div, "account")
	cli.ProjectMap = parseIdNameMap(div, "project")

	return nil
}

//解析HTML文档生成科目Map
func parseCategoryMap(doc *goquery.Selection) map[int]Category {
	categoryMap := map[int]Category{}
	anchors := doc.Find("div#panel-category a")
	for i := range anchors.Nodes {
		anchor := anchors.Eq(i)

		var categoryType int

		idStr, _ := anchor.Attr("id")
		idStr = strings.TrimSuffix(idStr, "-a")
		if strings.HasPrefix(idStr, "cCat-out-") {
			categoryType = TranTypePayout
			idStr = strings.TrimPrefix(idStr, "cCat-out-")
		} else if strings.HasPrefix(idStr, "cCat-in-") {
			categoryType = TranTypeIncome
			idStr = strings.TrimPrefix(idStr, "cCat-in-")
		} else {
			continue
		}

		id, _ := strconv.Atoi(idStr)

		category := Category{
			Type:   categoryType,
			IsSub:  !anchor.HasClass("ctit"),
			IdName: IdName{Id: id, Name: anchor.Text()},
		}
		categoryMap[id] = category
	}

	return categoryMap
}

//解析HTML文档生成类别Map
func parseIdNameMap(doc *goquery.Selection, zone string) map[int]IdName {
	prefix := "c" + strings.Title(zone[:3]) + "-"

	idNameMap := map[int]IdName{}
	anchors := doc.Find("div#panel-" + zone + " a")
	for i := range anchors.Nodes {
		anchor := anchors.Eq(i)

		idStr, _ := anchor.Attr("id")
		if idStr == prefix+"a" {
			continue
		}

		idStr = strings.TrimSuffix(idStr, "-a")
		idStr = strings.TrimPrefix(idStr, prefix)
		id, _ := strconv.Atoi(idStr)

		idNameMap[id] = IdName{Id: id, Name: anchor.Text()}
	}

	return idNameMap
}