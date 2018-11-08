package anilist

import (
	"context"
	"errors"

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
	Type       string
	MaxResults int
}

var (
	client *graphql.Client

	mediaIDQuery     string
	mediaTitleQuery  string
	mediaPersonQuery string
	mediaStudioQuery string
)

const (
	ANIME MediaType = 0
	MANGA MediaType = 1
)

func init() {
	client = graphql.NewClient("https://graphql.anilist.co/")

	mediaIDQuery = `
      query ($id: Int!) {
        Media(id: $id) {
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
        }
      }
	`

	mediaTitleQuery = `
      query ($search: String!, $max: Int!, $type: MediaType) {
				Page(page: 1, perPage: $max) {
					pageInfo {
						total
					}
					media(search: $search, type: $type, sort:POPULARITY_DESC) {
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
								isMain
								node {
									name
									siteUrl
								}
							}
						}
						staff {
							edges {
								role
								node{
									siteUrl
									name{
										first
										last
									}
								}
							}
						}
					}
				}
			}
	`

}

func MediaFromID(ctx context.Context, id int) (Media, error) {
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
	mediaQuery.Var("search", query.Title)
	mediaQuery.Var("max", query.MaxResults)
	if query.Type != "" {
		mediaQuery.Var("type", query.Type)
	}

	var res MediaPageResponse
	if err := client.Run(ctx, mediaQuery, &res); err != nil {
		return []Media{}, err
	}
	return res.Page.Media, nil
}

func MediaFromPersonQuery(ctx context.Context, query PersonQuery) ([]Media, error) {
	return []Media{}, nil
}

func MediaFromStudioQuery(ctx context.Context, query StudioQuery) ([]Media, error) {
	return []Media{}, nil
}

type MediaType int

func (t MediaType) Left() string {
	switch t {
	case ANIME:
		return "{"
	case MANGA:
		return "<"
	default:
		return " "
	}
}

func (t MediaType) Right() string {
	switch t {
	case ANIME:
		return "}"
	case MANGA:
		return ">"
	default:
		return " "
	}
}
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
func (t MediaType) MediaFromTitle(ctx context.Context, title string, maxResults int) ([]Media, error) {
	mediaQuery := MediaQuery{Title: title, Type: t.String(), MaxResults: maxResults}
	return MediaFromMediaQuery(ctx, mediaQuery)
}

func (t MediaType) MediaFromPersonName(ctx context.Context, name string, maxResults int) ([]Media, error) {
	personQuery := PersonQuery{Name: name, Type: t.String(), MaxResults: maxResults}
	return MediaFromPersonQuery(ctx, personQuery)
}

func (t MediaType) MediaFromPersonID(ctx context.Context, id int, maxResults int) ([]Media, error) {
	personQuery := PersonQuery{ID: id, Type: t.String(), MaxResults: maxResults}
	return MediaFromPersonQuery(ctx, personQuery)
}

func (t MediaType) MediaFromStudioName(ctx context.Context, name string, maxResults int) ([]Media, error) {
	studioQuery := StudioQuery{Name: name, Type: t.String(), MaxResults: maxResults}
	return MediaFromStudioQuery(ctx, studioQuery)
}

func (t MediaType) MediaFromStudioID(ctx context.Context, id int, maxResults int) ([]Media, error) {
	studioQuery := StudioQuery{ID: id, Type: t.String(), MaxResults: maxResults}
	return MediaFromStudioQuery(ctx, studioQuery)
}
