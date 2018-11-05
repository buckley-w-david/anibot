package anilist

import (
	"context"
	"errors"
	"fmt"

	"github.com/machinebox/graphql"
)

var (
	client *graphql.Client
)

func init() {
	client = graphql.NewClient("https://graphql.anilist.co/")
}

type MediaResponse struct {
	Media struct {
		SiteURL string `json:"siteUrl"`
		Title   struct {
			English string `json:"english,omitempty"`
			Romaji  string `json:"romaji,omitempty"`
		} `json:"title"`
		Description string `json:"description"`
		CoverImage  struct {
			ExtraLarge string `json:"extraLarge,omitempty"`
			Large      string `json:"large,omitempty"`
			Medium     string `json:"medium,omitempty"`
		} `json:"coverImage"`
		MediaType string `json:"type"`
		Format    string `json:"format"`
		Source    string `json:"source"`
		Studios   struct {
			Edges []struct {
				Node struct {
					Name    string `json:"name"`
					SiteURL string `json:"siteUrl"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"studios"`
		Staff struct {
			Edges []struct {
				Role string `json:"role"`
				Node struct {
					SiteURL string `json:"siteUrl"`
					Name    struct {
						First string `json:"first"`
						Last  string `json:"last"`
					} `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"staff"`
	} `json:"Media"`
}

type IDQuery struct {
	ID   int
	Type string
}

type TitleQuery struct {
	Title string
	Type  string
}

func (media MediaResponse) Director() (string, string, error) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Director" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return name, media.Media.Staff.Edges[i].Node.SiteURL, nil
		}
	}
	return "", "", errors.New("Unable to find director")
}

func (media MediaResponse) Creator() (string, string, error) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Original Creator" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return name, media.Media.Staff.Edges[i].Node.SiteURL, nil
		}
	}
	return "", "", errors.New("Unable to find creator")
}

func MediaFromID(ctx context.Context, query IDQuery) (MediaResponse, error) {
	idMediaQuery := graphql.NewRequest(`
      query ($id: Int!, $type: MediaType) {
        Media(id: $id, type: $type, sort: POPULARITY_DESC) {
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
	`)
	idMediaQuery.Var("id", query.ID)
	if query.Type != "" {
		idMediaQuery.Var("type", query.ID)
	}

	var res MediaResponse
	if err := client.Run(ctx, idMediaQuery, &res); err != nil {
		return MediaResponse{}, err
	}
	return res, nil
}

func MediaFromTitle(ctx context.Context, query TitleQuery) (MediaResponse, error) {
	titleMediaQuery := graphql.NewRequest(`
      query ($search: String!, $type: MediaType) {
        Media(search: $search, type: $type, sort: POPULARITY_DESC) {
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
	`)
	titleMediaQuery.Var("search", query.Title)
	if query.Type != "" {
		titleMediaQuery.Var("type", query.Type)
	}

	var res MediaResponse
	if err := client.Run(ctx, titleMediaQuery, &res); err != nil {
		return MediaResponse{}, err
	}
	return res, nil
}
