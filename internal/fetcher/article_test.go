package fetcher

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/ms-dw/configs"
	"github.com/pkg/errors"
)

// pass test
func TestFetchArticle(t *testing.T) {
	tests := []struct {
		url string
		err error
	}{
		{
			"https://www.dw.com/zh/%E5%BE%B7%E5%9B%BD%E4%BC%9A%E4%B8%8D%E4%BC%9A%E7%BB%A7%E7%BB%AD%E6%98%AF-%E6%AC%A7%E6%B4%B2%E5%A6%93%E9%99%A2/a-56587170",
			ErrTimeOverDays,
		},
		{
			"https://www.dw.com/zh/%E7%BE%8E%E6%8A%A5%E5%91%8A%E7%A7%B0%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%A4%96%E6%B3%84%E5%81%87%E8%AE%BE%E5%8F%AF%E4%BF%A1-%E5%B8%83%E6%9E%97%E8%82%AF%E8%B4%A8%E7%96%91%E7%A0%94%E7%A9%B6%E6%96%B9%E6%B3%95/a-57821645",
			nil,
		},
	}
	for _, tc := range tests {
		a := NewArticle()
		a, err := a.fetchArticle(tc.url)
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore old news pass test: ", tc.url)
			}
		} else {
			fmt.Println("pass test: ", a.Content)
		}
	}
}

func TestFetchTitle(t *testing.T) {
	tests := []struct {
		url   string
		title string
	}{
		{
			"https://www.dw.com/zh/%E5%BE%B7%E5%9B%BD%E4%BC%9A%E4%B8%8D%E4%BC%9A%E7%BB%A7%E7%BB%AD%E6%98%AF-%E6%AC%A7%E6%B4%B2%E5%A6%93%E9%99%A2/a-56587170",
			"德国会不会继续是 ″欧洲妓院″？",
		},
		{
			"https://www.dw.com/zh/%E7%BE%8E%E6%8A%A5%E5%91%8A%E7%A7%B0%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%A4%96%E6%B3%84%E5%81%87%E8%AE%BE%E5%8F%AF%E4%BF%A1-%E5%B8%83%E6%9E%97%E8%82%AF%E8%B4%A8%E7%96%91%E7%A0%94%E7%A9%B6%E6%96%B9%E6%B3%95/a-57821645",
			"美报告称实验室外泄假设可信 布林肯质疑研究方法",
		},
	}
	for _, tc := range tests {
		a := NewArticle()
		u, err := url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		a.U = u
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		got, err := a.fetchTitle()
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore pass test: ", tc.url)
			}
		} else {
			if tc.title != got {
				t.Errorf("\nwant: %s\n got: %s", tc.title, got)
			}
		}
	}

}

func TestFetchUpdateTime(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.dw.com/zh/%E5%BE%B7%E5%9B%BD%E4%BC%9A%E4%B8%8D%E4%BC%9A%E7%BB%A7%E7%BB%AD%E6%98%AF-%E6%AC%A7%E6%B4%B2%E5%A6%93%E9%99%A2/a-56587170",
			"2021-02-20 08:00:00 +0800 UTC",
		},
		{
			"https://www.dw.com/zh/%E7%BE%8E%E6%8A%A5%E5%91%8A%E7%A7%B0%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%A4%96%E6%B3%84%E5%81%87%E8%AE%BE%E5%8F%AF%E4%BF%A1-%E5%B8%83%E6%9E%97%E8%82%AF%E8%B4%A8%E7%96%91%E7%A0%94%E7%A9%B6%E6%96%B9%E6%B3%95/a-57821645",
			"2021-06-09 08:00:00 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		tt, err := a.fetchUpdateTime()
		if err != nil {
			t.Error(err)
		} else {
			ttt := tt.AsTime()
			got := shanghai(ttt)
			if got.String() != tc.want {
				t.Errorf("\nwant: %s\n got: %s", tc.want, got.String())
			}
		}
	}
}

func TestFetchContent(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.dw.com/zh/%E5%BE%B7%E5%9B%BD%E4%BC%9A%E4%B8%8D%E4%BC%9A%E7%BB%A7%E7%BB%AD%E6%98%AF-%E6%AC%A7%E6%B4%B2%E5%A6%93%E9%99%A2/a-56587170",
			"2021-02-20 08:00:00 +0800 UTC",
		},
		{
			"https://www.dw.com/zh/%E7%BE%8E%E6%8A%A5%E5%91%8A%E7%A7%B0%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%A4%96%E6%B3%84%E5%81%87%E8%AE%BE%E5%8F%AF%E4%BF%A1-%E5%B8%83%E6%9E%97%E8%82%AF%E8%B4%A8%E7%96%91%E7%A0%94%E7%A9%B6%E6%96%B9%E6%B3%95/a-57821645",
			"2021-06-09 08:00:00 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		c, err := a.fetchContent()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(c)
	}
}
