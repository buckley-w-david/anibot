package anilist

import (
	"context"
	"errors"
	"fmt"

	"github.com/machinebox/graphql"
)

var (
	client       *graphql.Client
	idMediaQuery *graphql.Request
)

func init() {
	idMediaQuery = graphql.NewRequest(`
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
		}`)

	client = graphql.NewClient("https://graphql.anilist.co/")
}

type MediaResponse struct {
	Media struct {
		SiteUrl string `json:"siteUrl"`
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
					SiteUrl string `json:"siteUrl"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"studios"`
		Staff struct {
			Edges []struct {
				Role string `json:"role"`
				Node struct {
					SiteUrl string `json:"siteUrl"`
					Name    struct {
						First string `json:"first"`
						Last  string `json:"last"`
					} `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"staff"`
	} `json:"Media"`
}

func (media MediaResponse) Director() (error, string, string) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Director" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return nil, name, media.Media.Staff.Edges[i].Node.SiteUrl
		}
	}
	return errors.New("Unable to find director"), "", ""
}

func (media MediaResponse) Creator() (error, string, string) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Original Creator" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return nil, name, media.Media.Staff.Edges[i].Node.SiteUrl
		}
	}
	return errors.New("Unable to find director"), "", ""
}

func MediaFromId(ctx context.Context, id int) (error, MediaResponse) {
	idMediaQuery.Var("id", id)

	var res MediaResponse
	if err := client.Run(ctx, idMediaQuery, &res); err != nil {
		fmt.Println("Error making GraphQL query", err)
		return err, MediaResponse{}
	}
	return nil, res
}
