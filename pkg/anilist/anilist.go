package anilist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	//"github.com/davecgh/go-spew/spew"
	"github.com/machinebox/graphql"
)

type Person struct {
	ID      int    `json:"id"`
	SiteURL string `json:"siteUrl"`
	Name    struct {
		First string `json:"first"`
		Last  string `json:"last"`
	}
}

type Studio struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	SiteURL string `json:"siteUrl"`
}

type Title struct {
	English string `json:"english,omitempty"`
	Romaji  string `json:"romaji,omitempty"`
}

type CoverImage struct {
	ExtraLarge string `json:"extraLarge,omitempty"`
	Large      string `json:"large,omitempty"`
	Medium     string `json:"medium,omitempty"`
}

type Media struct {
	SiteURL     string     `json:"siteUrl"`
	Title       Title      `json:"title"`
	Description string     `json:"description"`
	CoverImage  CoverImage `json:"coverImage"`
	MediaType   string     `json:"type"`
	Format      string     `json:"format"`
	Source      string     `json:"source"`
	Studios     struct {
		Edges []struct {
			Studio Studio `json:"node"`
		} `json:"edges"`
	} `json:"studios"`
	Staff struct {
		Edges []struct {
			Role   string `json:"role"`
			Person Person `json:"node"`
		} `json:"edges"`
	} `json:"staff"`
}

type MediaResponse struct {
	Media Media `json:"Media"`
}

type MediaPageResponse struct {
	Page struct {
		PageInfo struct {
			Total int `json:"total"`
		} `json:"pageInfo"`
		Media []Media `json:"media"`
	}
}

type StaffMediaResponse struct {
	Staff struct {
		StaffMedia struct {
			Nodes []struct {
				ID int `json:"id"`
			} `json:"nodes"`
		} `json:"staffMedia"`
	} `json:"Staff"`
}

type StudioMediaResponse struct {
	Studio struct {
		Media struct {
			Nodes []struct {
				ID int `json:"id"`
			} `json:"nodes"`
		} `json:"media"`
	} `json:"Studio"`
}

func (media Media) Director() (Person, error) {
	for i := range media.Staff.Edges {
		if media.Staff.Edges[i].Role == "Director" {
			return media.Staff.Edges[i].Person, nil
		}
	}
	return Person{}, errors.New("Unable to find director")
}

func (media Media) Creator() (Person, error) {
	for i := range media.Staff.Edges {
		if media.Staff.Edges[i].Role == "Original Creator" {
			return media.Staff.Edges[i].Person, nil
		}
	}
	return Person{}, errors.New("Unable to find creator")
}

type MediaQuery struct {
	Title      string
	ID         int
	Type       string
	Sort       []string
	MaxResults int
}

type PersonQuery struct {
	Name       string
	ID         int
	Type       string
	MaxResults int
}

type StudioQuery struct {
	Name       string
	ID         int
	MaxResults int
}

var (
	client *graphql.Client

	mediaIDQuery    string
	mediaTitleQuery string

	mediaPersonQuery string

	mediaStudioQuery string
)

const (
	ANIME MediaType = 0
	MANGA MediaType = 1
)

func init() {
	client = graphql.NewClient("https://graphql.anilist.co/")

	media := `
      siteUrl
      title{
        english
        romaji
      }
      description(asHtml: false)
      coverImage {
        extraLarge
        large
        medium
      }
      type
      format
      source
      studios {
        edges {
          node {
            id
            name
            siteUrl
          }
        }
      }
      staff {
        edges {
          role
          node{
            id
            siteUrl
            name{
              first
              last
            }
          }
        }
      }
	`

	mediaIDQuery = fmt.Sprintf(`query ($id: Int!) { Media(id: $id) { %s } }`, media)

	mediaTitleQuery = fmt.Sprintf(`
      query ($search: String, $id: Int, $max: Int!, $type: MediaType, $sort: [MediaSort]) {
        Page(page: 1, perPage: $max) {
          pageInfo {
            total
          }
          media(search: $search, id: $id, type: $type, sort: $sort) {
            %s
          }
        }
      }
    `, media)

	mediaPersonQuery = `
	  query ($id: Int, $search: String, $max: Int!, $type: MediaType) {
        Staff(id: $id, search: $search) {
          staffMedia(sort:POPULARITY_DESC, type: $type, page: 1, perPage: $max) {
            nodes {
              id
            }
          } 
        }
      }
		`

	mediaStudioQuery = `
    query ($id: Int, $search: String, $max: Int!) {
      Studio(id: $id, search: $search) {
        media(sort:POPULARITY_DESC, page: 1, perPage: $max) {
          nodes{
            id
          }
        }
      }
    }
    `
}

func MediaFromMediaID(ctx context.Context, id int) (Media, error) {
	idMediaQuery := graphql.NewRequest(mediaIDQuery)
	idMediaQuery.Var("id", id)

	var res MediaResponse
	if err := client.Run(ctx, idMediaQuery, &res); err != nil {
		return Media{}, err
	}
	return res.Media, nil
}

func MediaFromMediaQuery(ctx context.Context, query MediaQuery) ([]Media, error) {
	mediaQuery := graphql.NewRequest(mediaTitleQuery)
	mediaQuery.Var("max", query.MaxResults)
	if query.Title != "" {
		mediaQuery.Var("search", query.Title)
	} else if query.ID != 0 {
		mediaQuery.Var("id", query.ID)
	}
	if query.Type != "" {
		mediaQuery.Var("type", query.Type)
	}
	if len(query.Sort) > 0 {
		mediaQuery.Var("sort", query.Sort)
	}

	var res MediaPageResponse
	if err := client.Run(ctx, mediaQuery, &res); err != nil {
		return []Media{}, err
	}
	return res.Page.Media, nil
}

func MediaFromPersonQuery(ctx context.Context, query PersonQuery) (response []Media, err error) {
	mediaQuery := graphql.NewRequest(mediaPersonQuery)
	if query.Name != "" {
		mediaQuery.Var("name", query.Name)
	} else if query.ID != 0 {
		mediaQuery.Var("id", query.ID)
	} else {
		return []Media{}, errors.New("Neither ID or Name set in PersonQuery")
	}

	if query.Type != "" {
		mediaQuery.Var("type", query.Type)
	}
	mediaQuery.Var("max", query.MaxResults)

	var res StaffMediaResponse
	if err := client.Run(ctx, mediaQuery, &res); err != nil {
		return []Media{}, err
	}
	for i := 0; i < len(res.Staff.StaffMedia.Nodes); i++ {
		media, err := MediaFromMediaID(ctx, res.Staff.StaffMedia.Nodes[i].ID)
		if err == nil {
			response = append(response, media)
		}
	}
	return
}

func MediaFromStudioQuery(ctx context.Context, query StudioQuery) (response []Media, err error) {
	mediaQuery := graphql.NewRequest(mediaStudioQuery)
	if query.Name != "" {
		mediaQuery.Var("name", query.Name)
	} else if query.ID != 0 {
		mediaQuery.Var("id", query.ID)
	} else {
		return []Media{}, errors.New("Neither ID or Name set in PersonQuery")
	}
	mediaQuery.Var("max", query.MaxResults)

	var res StudioMediaResponse
	if err := client.Run(ctx, mediaQuery, &res); err != nil {
		return []Media{}, err
	}
	for i := 0; i < len(res.Studio.Media.Nodes); i++ {
		media, err := MediaFromMediaID(ctx, res.Studio.Media.Nodes[i].ID)
		if err == nil {
			response = append(response, media)
		}
	}
	return

}

type MediaType int

func (t MediaType) String() string {
	switch t {
	case ANIME:
		return "ANIME"
	case MANGA:
		return "MANGA"
	default:
		return ""
	}
}

func MediaFromTitle(ctx context.Context, title string, maxResults int) ([]Media, error) {
	mediaQuery := MediaQuery{Title: title, MaxResults: maxResults}
	return MediaFromMediaQuery(ctx, mediaQuery)
}

func MediaFromPersonName(ctx context.Context, name string, maxResults int) ([]Media, error) {
	personQuery := PersonQuery{Name: name, MaxResults: maxResults}
	return MediaFromPersonQuery(ctx, personQuery)
}

func MediaFromPersonID(ctx context.Context, id int, maxResults int) ([]Media, error) {
	personQuery := PersonQuery{ID: id, MaxResults: maxResults}
	return MediaFromPersonQuery(ctx, personQuery)
}

func MediaFromStudioName(ctx context.Context, name string, maxResults int) ([]Media, error) {
	studioQuery := StudioQuery{Name: name, MaxResults: maxResults}
	return MediaFromStudioQuery(ctx, studioQuery)
}

func MediaFromStudioID(ctx context.Context, id int, maxResults int) ([]Media, error) {
	studioQuery := StudioQuery{ID: id, MaxResults: maxResults}
	return MediaFromStudioQuery(ctx, studioQuery)
}

// Execute is for specialized more specific queries that clients may want to perform that the library does not
// explicitly support. Try not to use this if at all possible.
func Execute(ctx context.Context, query string, vars map[string]interface{}) (map[string]*json.RawMessage, error) {
	mediaQuery := graphql.NewRequest(query)
	for k, v := range vars {
		mediaQuery.Var(k, v)
	}

	var res map[string]*json.RawMessage
	if err := client.Run(ctx, mediaQuery, &res); err != nil {
		return map[string]*json.RawMessage{}, err
	}
	return res, nil
}
