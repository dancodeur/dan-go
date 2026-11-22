// Package open5e est un client pour l'API Open5e (monstres D&D 5e).
package open5e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DefaultBaseURL est l'adresse de l'API publique.
const DefaultBaseURL = "https://api.open5e.com/v1"

const (
	requestTimeout = 10 * time.Second
	pageSize       = 10
)

// ErrNotFound est renvoyée quand l'API ne connaît pas le monstre demandé.
var ErrNotFound = errors.New("monstre introuvable")

// Client interroge l'API et garde les réponses dans le cache.
type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      *Cache
}

func NewClient(baseURL string, cache *Cache) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{Timeout: requestTimeout},
		cache:      cache,
	}
}

// ListMonsters renvoie une page de monstres. Avec une adresse vide, c'est la
// première page ; sinon on passe le Next ou le Previous de la page précédente.
func (c *Client) ListMonsters(pageURL string) (MonsterPage, error) {
	if pageURL == "" {
		pageURL = fmt.Sprintf("%s/monsters/?limit=%d", c.baseURL, pageSize)
	}

	var page MonsterPage
	if err := c.getJSON(pageURL, &page); err != nil {
		return MonsterPage{}, fmt.Errorf("liste des monstres : %w", err)
	}
	return page, nil
}

// GetMonster cherche un monstre par son slug ("giant-spider") ou son nom
// ("giant spider"). Si le slug n'existe pas, on essaie la recherche de l'API.
func (c *Client) GetMonster(name string) (Monster, error) {
	slug := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")

	monster, err := c.getMonsterBySlug(slug)
	if err == nil {
		return monster, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Monster{}, err
	}
	return c.searchMonster(name)
}

// Prefetch charge plusieurs monstres en parallèle pour remplir le cache.
// Elle attend que tous les téléchargements soient finis.
func (c *Client) Prefetch(names []string) error {
	errs := make([]error, len(names))

	var group sync.WaitGroup
	for index, name := range names {
		group.Add(1)
		go func() {
			defer group.Done()
			_, errs[index] = c.GetMonster(name)
		}()
	}
	group.Wait()

	return errors.Join(errs...)
}

func (c *Client) getMonsterBySlug(slug string) (Monster, error) {
	var monster Monster
	requestURL := c.baseURL + "/monsters/" + url.PathEscape(slug) + "/"
	if err := c.getJSON(requestURL, &monster); err != nil {
		return Monster{}, fmt.Errorf("récupération du monstre %q : %w", slug, err)
	}
	return monster, nil
}

func (c *Client) searchMonster(name string) (Monster, error) {
	var page MonsterPage
	requestURL := c.baseURL + "/monsters/?search=" + url.QueryEscape(name)
	if err := c.getJSON(requestURL, &page); err != nil {
		return Monster{}, fmt.Errorf("recherche du monstre %q : %w", name, err)
	}
	if len(page.Results) == 0 {
		return Monster{}, fmt.Errorf("monstre %q : %w", name, ErrNotFound)
	}

	// On préfère le nom exact, sinon le premier résultat.
	for _, monster := range page.Results {
		if strings.EqualFold(monster.Name, name) {
			return monster, nil
		}
	}
	return page.Results[0], nil
}

// getJSON télécharge une adresse (ou la prend dans le cache) et décode le JSON.
func (c *Client) getJSON(requestURL string, target any) error {
	data, err := c.fetch(requestURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("lecture de la réponse de %s : %w", requestURL, err)
	}
	return nil
}

func (c *Client) fetch(requestURL string) ([]byte, error) {
	if data, found := c.cache.Get(requestURL); found {
		return data, nil
	}

	// L'erreur de http.Client contient déjà l'adresse demandée.
	response, err := c.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("API injoignable : %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s: %w", requestURL, ErrNotFound)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requête %s : statut inattendu %s", requestURL, response.Status)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture de la réponse de %s : %w", requestURL, err)
	}

	c.cache.Set(requestURL, data)
	return data, nil
}
