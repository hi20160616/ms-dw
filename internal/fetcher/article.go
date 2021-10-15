package fetcher

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/gears"
	"github.com/hi20160616/ms-dw/configs"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Article struct {
	Id            string
	Title         string
	Content       string
	WebsiteId     string
	WebsiteDomain string
	WebsiteTitle  string
	UpdateTime    *timestamppb.Timestamp
	U             *url.URL
	raw           []byte
	doc           *html.Node
}

func NewArticle() *Article {
	return &Article{
		WebsiteDomain: configs.Data.MS["dw"].Domain,
		WebsiteTitle:  configs.Data.MS["dw"].Title,
		WebsiteId:     fmt.Sprintf("%x", md5.Sum([]byte(configs.Data.MS["dw"].Domain))),
	}
}

// List get all articles from database
func (a *Article) List() ([]*Article, error) {
	return load()
}

// Get read database and return the data by rawurl.
func (a *Article) Get(id string) (*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		if a.Id == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("[%s] no article with id: %s",
		configs.Data.MS["dw"].Title, id)
}

func (a *Article) Search(keyword ...string) ([]*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	as2 := []*Article{}
	for _, a := range as {
		for _, v := range keyword {
			v = strings.ToLower(strings.TrimSpace(v))
			switch {
			case a.Id == v:
				as2 = append(as2, a)
			case a.WebsiteId == v:
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Title), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Content), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteDomain), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteTitle), v):
				as2 = append(as2, a)
			}
		}
	}
	return as2, nil
}

type ByUpdateTime []*Article

func (u ByUpdateTime) Len() int      { return len(u) }
func (u ByUpdateTime) Swap(i, j int) { u[i], u[j] = u[j], u[i] }
func (u ByUpdateTime) Less(i, j int) bool {
	return u[i].UpdateTime.AsTime().Before(u[j].UpdateTime.AsTime())
}

var timeout = func() time.Duration {
	t, err := time.ParseDuration(configs.Data.MS["dw"].Timeout)
	if err != nil {
		log.Printf("[%s] timeout init error: %v", configs.Data.MS["dw"].Title, err)
		return time.Duration(1 * time.Minute)
	}
	return t
}()

// fetchArticle fetch article by rawurl
func (a *Article) fetchArticle(rawurl string) (*Article, error) {
	var err error
	a.U, err = url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// Dail
	a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
	if err != nil {
		return nil, err
	}

	a.Id = fmt.Sprintf("%x", md5.Sum([]byte(rawurl)))

	a.Title, err = a.fetchTitle()
	if err != nil {
		return nil, err
	}

	a.UpdateTime, err = a.fetchUpdateTime()
	if err != nil {
		return nil, err
	}

	// content should be the last step to fetch
	a.Content, err = a.fetchContent()
	if err != nil {
		return nil, err
	}

	a.Content, err = a.fmtContent(a.Content)
	if err != nil {
		return nil, err
	}
	return a, nil

}

func (a *Article) fetchTitle() (string, error) {
	n := exhtml.ElementsByTag(a.doc, "title")
	if n == nil || len(n) == 0 {
		return "", fmt.Errorf("[%s] there is no element <title>: %s",
			configs.Data.MS["dw"].Title, a.U.String())
	}
	title := n[0].FirstChild.Data
	title = strings.TrimSpace(title[:strings.Index(title, "|")])
	gears.ReplaceIllegalChar(&title)
	return title, nil
}

func (a *Article) fetchUpdateTime() (*timestamppb.Timestamp, error) {
	if a.raw == nil {
		return nil, fmt.Errorf("[%s] fetchUpdateTime: raw is nil: %s", configs.Data.MS["dw"].Title, a.U.String())
	}
	re := regexp.MustCompile(`articleChangeDateShort: "(\d*?)",`)
	rs := re.FindAllSubmatch(a.raw, -1)
	if len(rs) <= 0 {
		return nil, fmt.Errorf("[%s] fetchUpdateTime regex match nothing: %s",
			configs.Data.MS["dw"].Title, a.U.String())
	}
	if len(rs[0]) <= 1 {
		return nil, fmt.Errorf("[%s] fetchUpdateTime regex match nothing: %s",
			configs.Data.MS["dw"].Title, a.U.String())
	}
	t, err := time.Parse("20060102", string(rs[0][1]))
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}

func shanghai(t time.Time) time.Time {
	loc := time.FixedZone("UTC", 8*60*60)
	return t.In(loc)
}

func (a *Article) fetchContent() (string, error) {
	if a.doc == nil || a.raw == nil {
		return "", errors.Errorf("[%s] fetchContent: doc or raw is nil: %s", configs.Data.MS["dw"].Title, a.U.String())
	}
	body := ""
	// Fetch content nodes
	// Fetch summary
	re := regexp.MustCompile(`<p class="intro">(.*?)</p>`)
	raw := bytes.ReplaceAll(a.raw, []byte("\n"), []byte(""))
	rs := re.FindAllSubmatch(raw, -1)
	if rs == nil {
		return "", errors.New("dw: dw: intro match nothing.")
	}
	if intro := string(rs[0][1]); intro != "" {
		body += "> " + intro + "  \n\n" // if intro exist, append to body
	}

	// Fetch content
	nodes := exhtml.ElementsByTagAndClass(a.doc, "div", "longText")
	if len(nodes) == 0 {
		return "", fmt.Errorf("[%s] nodes fetch error from: %s",
			configs.Data.MS["dw"].Title, a.U.String())
	}

	x := nodes[0]
	if x != nil {
		x = x.FirstChild
		if x != nil {
			x = x.NextSibling
			if x != nil {
				if len(x.Attr) > 0 {
					if x.Attr[0].Val == "col1" {
						nodes[0].RemoveChild(nodes[0].FirstChild.NextSibling)
					}
				}
			}
		}
	}

	spanMerge := func(n *html.Node) []*html.Node {
		spans := exhtml.ElementsByTag(n, "span")
		for _, span := range spans {
			if span.FirstChild != nil {
				body += span.FirstChild.Data
				if span.FirstChild.Data != span.LastChild.Data {
					body += span.LastChild.Data
				}
			}
		}
		return spans
	}

	plist := exhtml.ElementsByTag(nodes[0], "p", "h2")
	for _, v := range plist {
		if v.FirstChild == nil {
			continue
		} else {
			switch v.Data {
			case "h2":
				body += "\n**"
				if ss := spanMerge(v); len(ss) == 0 {
					body += v.FirstChild.Data
				}
				body += "**   \n"
			case "p":
				if ss := spanMerge(v); len(ss) == 0 {
					body += v.FirstChild.Data
				}
				body += "  \n"
			default:
				body += v.FirstChild.Data + "  \n"
			}
		}
	}
	rp := strings.NewReplacer("strong  \n", "", "em  \n", "", "**\n**   \n", "")
	rp.Replace(body)
	return body, nil
}

func (a *Article) fmtContent(body string) (string, error) {
	var err error
	title := "# " + a.Title + "\n\n"
	lastupdate := a.UpdateTime.AsTime().Format(time.RFC3339)
	webTitle := fmt.Sprintf(" @ [%s](/list/?v=%[1]s): [%[2]s](http://%[2]s)", a.WebsiteTitle, a.WebsiteDomain)
	u, err := url.QueryUnescape(a.U.String())
	if err != nil {
		u = a.U.String() + "\n\nunescape url error:\n" + err.Error()
	}

	body = title +
		"LastUpdate: " + lastupdate +
		webTitle + "\n\n" +
		"---\n" +
		body + "\n\n" +
		"原地址：" + fmt.Sprintf("[%s](%[1]s)", u)
	return body, nil
}
