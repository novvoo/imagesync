package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/compression"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ======================
// Registry API 工具函数
// ======================

func listProjectsHarbor(urlBase, user, password string) ([]string, error) {
	var projects []string
	page := 1
	client := &http.Client{Timeout: 10 * time.Second}

	for {
		apiURL := fmt.Sprintf("%s/api/v2.0/projects?page=%d&page_size=100", urlBase, page)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("projects request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("projects API error %d: %s", resp.StatusCode, string(body))
		}

		var projList []struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&projList); err != nil {
			return nil, fmt.Errorf("parse projects failed: %w", err)
		}

		if len(projList) == 0 {
			break
		}

		for _, p := range projList {
			projects = append(projects, p.Name)
		}

		page++
	}

	return projects, nil
}

func listReposInProjectHarbor(urlBase, user, password, project string) ([]string, error) {
	var repos []string
	page := 1
	client := &http.Client{Timeout: 10 * time.Second}

	for {
		apiURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page=%d&page_size=100", urlBase, url.PathEscape(project), page)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("repos request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("repos API error %d: %s", resp.StatusCode, string(body))
		}

		var repoList []struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&repoList); err != nil {
			return nil, fmt.Errorf("parse repos failed: %w", err)
		}

		if len(repoList) == 0 {
			break
		}

		for _, r := range repoList {
			repos = append(repos, r.Name)
		}

		page++
	}

	return repos, nil
}

func listRepositoriesHarbor(urlBase, user, password, prefix string) ([]string, error) {
	var repos []string

	if prefix == "" {
		projects, err := listProjectsHarbor(urlBase, user, password)
		if err != nil {
			return nil, err
		}
		for _, project := range projects {
			projRepos, err := listReposInProjectHarbor(urlBase, user, password, project)
			if err != nil {
				return nil, err
			}
			for _, r := range projRepos {
				repos = append(repos, r)
			}
		}
	} else {
		prefix = strings.TrimSuffix(prefix, "/")
		parts := strings.Split(prefix, "/")
		project := parts[0]
		subPrefix := ""
		if len(parts) > 1 {
			subPrefix = strings.Join(parts[1:], "/") + "/"
		}
		projRepos, err := listReposInProjectHarbor(urlBase, user, password, project)
		if err != nil {
			return nil, err
		}
		prefixWithSlash := prefix + "/"
		for _, r := range projRepos {
			if subPrefix == "" || strings.HasPrefix(r, prefixWithSlash) {
				repos = append(repos, r)
			}
		}
	}

	return repos, nil
}

func listRepositoriesRegistry(urlBase, user, password, prefix string) ([]string, error) {
	host := strings.TrimPrefix(strings.TrimPrefix(urlBase, "https://"), "http://")
	var repos []string
	nextURL := fmt.Sprintf("https://%s/v2/_catalog?n=100", host)
	client := &http.Client{Timeout: 10 * time.Second}

	prefix = strings.TrimSuffix(prefix, "/")
	if prefix != "" {
		prefix += "/"
	}

	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("catalog request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("catalog API error %d: %s", resp.StatusCode, string(body))
		}

		var cat struct {
			Repositories []string `json:"repositories"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
			return nil, fmt.Errorf("parse catalog failed: %w", err)
		}

		for _, r := range cat.Repositories {
			if prefix == "" || strings.HasPrefix(r, prefix) {
				repos = append(repos, r)
			}
		}

		nextURL = ""
		link := resp.Header.Get("Link")
		if link != "" && strings.Contains(link, `rel="next"`) {
			start := strings.Index(link, "<")
			end := strings.Index(link, ">")
			if start != -1 && end != -1 {
				nextURL = "https://" + host + link[start+1:end]
			}
		}
	}

	return repos, nil
}

func listRepositories(urlBase, user, password, prefix, typ string) ([]string, error) {
	if typ == "harbor" {
		return listRepositoriesHarbor(urlBase, user, password, prefix)
	} else if typ == "acr" {
		return listRepositoriesRegistry(urlBase, user, password, prefix)
	}
	return nil, fmt.Errorf("unsupported type: %s", typ)
}

func listTags(urlBase, user, password, repoPath, typ string) ([]string, error) {
	host := strings.TrimPrefix(strings.TrimPrefix(urlBase, "https://"), "http://")
	baseURL := fmt.Sprintf("https://%s/v2/%s/tags/list?n=100", host, url.PathEscape(repoPath))
	var tags []string
	nextURL := baseURL
	client := &http.Client{Timeout: 10 * time.Second}

	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("tags request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("tags API error %d: %s", resp.StatusCode, string(body))
		}

		var tagStruct struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tagStruct); err != nil {
			return nil, fmt.Errorf("parse tags failed: %w", err)
		}

		tags = append(tags, tagStruct.Tags...)

		nextURL = ""
		link := resp.Header.Get("Link")
		if link != "" && strings.Contains(link, `rel="next"`) {
			start := strings.Index(link, "<")
			end := strings.Index(link, ">")
			if start != -1 && end != -1 {
				nextURL = "https://" + host + link[start+1:end]
			}
		}
	}

	return tags, nil
}

// 从 SRC 获取最新 tag
func getLatestTag(urlBase, user, password, repoPath, typ string) (string, error) {
	if typ == "harbor" {
		parts := strings.SplitN(repoPath, "/", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid repo path: must be project/repository, got %s", repoPath)
		}
		projectName := parts[0]
		repositoryName := parts[1]

		apiURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page=1&page_size=1&with_tag=true&sort=creation_time%%20desc",
			urlBase, url.PathEscape(projectName), url.PathEscape(url.PathEscape(repositoryName)))

		req, _ := http.NewRequest("GET", apiURL, nil)
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("Harbor API error %d: %s", resp.StatusCode, string(body))
		}

		var artifacts []struct {
			Tags []struct {
				Name string `json:"name"`
			} `json:"tags"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
			return "", fmt.Errorf("parse response failed: %w", err)
		}

		if len(artifacts) == 0 || len(artifacts[0].Tags) == 0 {
			return "", fmt.Errorf("no tags found for %s", repoPath)
		}

		return artifacts[0].Tags[0].Name, nil
	} else if typ == "acr" {
		apiURL := fmt.Sprintf("%s/acr/v1/%s/_tags?orderby=timedesc&n=1", urlBase, url.PathEscape(repoPath))

		req, _ := http.NewRequest("GET", apiURL, nil)
		req.SetBasicAuth(user, password)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("ACR API error %d: %s", resp.StatusCode, string(body))
		}

		var tagResp struct {
			TagsAttributes []struct {
				Name string `json:"name"`
			} `json:"tagsAttributes"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tagResp); err != nil {
			return "", fmt.Errorf("parse response failed: %w", err)
		}

		if len(tagResp.TagsAttributes) == 0 {
			return "", fmt.Errorf("no tags found for %s", repoPath)
		}

		return tagResp.TagsAttributes[0].Name, nil
	}
	return "", fmt.Errorf("unsupported type: %s", typ)
}

// 检查 DST Registry（支持 Harbor 或 ACR）是否已存在该 tag
// 使用 Docker Registry v2 API 的 manifest endpoint，兼容 ACR 和 Harbor
func imageExistsInDst(urlBase, user, password, repoPath, tag string) (bool, error) {
	host := strings.TrimPrefix(strings.TrimPrefix(urlBase, "https://"), "http://")
	apiURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", host, url.PathEscape(repoPath), url.PathEscape(tag))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("create request failed: %w", err)
	}
	req.SetBasicAuth(user, password)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("check existence failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("check existence API error %d: %s", resp.StatusCode, string(body))
	}
}

// ======================
// 同步逻辑
// ======================
func syncImageWithZstd(ctx context.Context, srcRef, dstRef string, srcSysCtx, dstSysCtx *types.SystemContext) error {
	dstSysCtx.CompressionFormat = &compression.Zstd
	dstSysCtx.CompressionLevel = nil

	srcImgRef, err := alltransports.ParseImageName("docker://" + srcRef)
	if err != nil {
		return fmt.Errorf("parse src ref: %w", err)
	}
	dstImgRef, err := alltransports.ParseImageName("docker://" + dstRef)
	if err != nil {
		return fmt.Errorf("parse dst ref: %w", err)
	}

	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()},
	}
	pc, err := signature.NewPolicyContext(policy)
	if err != nil {
		return fmt.Errorf("creating policy context: %w", err)
	}
	defer pc.Destroy()

	srcImageSrc, err := srcImgRef.NewImageSource(ctx, srcSysCtx)
	if err != nil {
		return fmt.Errorf("create source image: %w", err)
	}
	defer srcImageSrc.Close()

	_, manifestType, err := srcImageSrc.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("get manifest: %w", err)
	}

	isMultiArch := manifest.MIMETypeIsMultiImage(manifestType)

	comp := copy.OptionCompressionVariant{
		Algorithm: compression.Zstd,
		Level:     nil,
	}

	copyOpts := &copy.Options{
		SourceCtx:              srcSysCtx,
		DestinationCtx:         dstSysCtx,
		ForceCompressionFormat: true,
	}

	if isMultiArch {
		copyOpts.EnsureCompressionVariantsExist = []copy.OptionCompressionVariant{comp}
	}

	_, err = copy.Image(ctx, pc, dstImgRef, srcImgRef, copyOpts)
	if err != nil {
		return fmt.Errorf("copy image with zstd: %w", err)
	}

	return nil
}

// ======================
// 数据库操作
// ======================

func initDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS image_records`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE image_records (
			id SERIAL PRIMARY KEY,
			repo_path TEXT NOT NULL,
			tag TEXT NOT NULL,
			src_oci TEXT NOT NULL,
			dst_oci TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE (repo_path, tag)
		);`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func storeRecord(db *sql.DB, repoPath, tag, srcOCI, dstOCI string) error {
	query := `
		INSERT INTO image_records (repo_path, tag, src_oci, dst_oci, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (repo_path, tag)
		DO UPDATE SET src_oci = $3, dst_oci = $4, updated_at = NOW();`
	_, err := db.Exec(query, repoPath, tag, srcOCI, dstOCI)
	return err
}

// ======================
// 主函数
// ======================

func main() {
	cfg := loadConfig()

	// Parse command line
	var inputPath string
	if len(os.Args) > 1 {
		inputPath = os.Args[1]
	}
	if inputPath == "" {
		inputPath = "/library"
	}

	// Initialize DB
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Database)
	db, err := initDB(dsn)
	if err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	// Parse path
	path := strings.TrimPrefix(inputPath, "/")
	var specificTag string
	colonIdx := strings.LastIndex(path, ":")
	if colonIdx != -1 {
		specificTag = path[colonIdx+1:]
		path = path[:colonIdx]
	}

	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// Get list of repos
	repos, err := listRepositories(cfg.Src.URLBase, cfg.Src.User, cfg.Src.Password, prefix, cfg.Src.Type)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v", err)
	}

	var isSpecific bool
	if len(repos) == 0 && path != "" {
		prefix = strings.TrimSuffix(prefix, "/")
		repos = []string{prefix}
		isSpecific = true
	}

	srcHost := strings.TrimPrefix(strings.TrimPrefix(cfg.Src.URLBase, "https://"), "http://")
	dstHost := strings.TrimPrefix(strings.TrimPrefix(cfg.Dst.URLBase, "https://"), "http://")
	srcSysCtx := &types.SystemContext{
		DockerInsecureSkipTLSVerify: types.OptionalBoolFalse,
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: cfg.Src.User,
			Password: cfg.Src.Password,
		},
	}
	dstSysCtx := &types.SystemContext{
		DockerInsecureSkipTLSVerify: types.OptionalBoolFalse,
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: cfg.Dst.User,
			Password: cfg.Dst.Password,
		},
	}

	for _, repoPath := range repos {
		var tags []string
		if specificTag != "" {
			tags = []string{specificTag}
		} else if isSpecific {
			tags, err = listTags(cfg.Src.URLBase, cfg.Src.User, cfg.Src.Password, repoPath, cfg.Src.Type)
			if err != nil {
				log.Printf("❌ Failed to list tags for %s: %v", repoPath, err)
				continue
			}
		} else {
			tag, err := getLatestTag(cfg.Src.URLBase, cfg.Src.User, cfg.Src.Password, repoPath, cfg.Src.Type)
			if err != nil {
				log.Printf("⚠️ Skipping %s: failed to get latest tag: %v", repoPath, err)
				continue
			}
			tags = []string{tag}
		}

		for _, tag := range tags {
			srcOCI := fmt.Sprintf("%s/%s:%s", srcHost, repoPath, tag)
			dstOCI := fmt.Sprintf("%s/%s:%s", dstHost, repoPath, tag)

			log.Printf("📍 Processing %s:%s", repoPath, tag)

			exists, err := imageExistsInDst(cfg.Dst.URLBase, cfg.Dst.User, cfg.Dst.Password, repoPath, tag)
			if err != nil {
				log.Printf("❌ Failed to check existence for %s:%s: %v", repoPath, tag, err)
				continue
			}

			if exists {
				log.Printf("✅ Already exists: %s:%s", repoPath, tag)
			} else {
				log.Printf("🔄 Syncing %s:%s", repoPath, tag)

				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
				defer cancel()

				err = syncImageWithZstd(ctx, srcOCI, dstOCI, srcSysCtx, dstSysCtx)
				if err != nil {
					log.Printf("❌ Sync failed for %s:%s: %v", repoPath, tag, err)
					continue
				}
				log.Printf("✅ Synced %s:%s", repoPath, tag)
			}

			// Store record
			if err := storeRecord(db, repoPath, tag, srcOCI, dstOCI); err != nil {
				log.Printf("❌ Failed to store record for %s:%s: %v", repoPath, tag, err)
			}
		}
	}

	log.Println("🎉 All sync jobs completed!")
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not loaded, using system environment variables")
	}
}
