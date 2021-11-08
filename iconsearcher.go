package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"net/http"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Cfg struct {
	IconsFolder string `yaml:"icons_folder"`
}

type Dmi struct {
	Name string `json:"dmi"`
	Url  string `json:"url"`
	Icon []Icon `json:"icons"`
}

type Icon struct {
	Name     string `json:"name"`
	DmiPath  string `json:"dmi"`
	filepath string
	dmi      string
}

var icons map[string]Icon
var icons_ordered_by_dmi map[string][]*Icon

var AppConfig *Cfg

func orderbydmi(folder string, icon *Icon) {
	ico := icons_ordered_by_dmi[folder]
	ico = append(ico, icon)
	icons_ordered_by_dmi[folder] = ico
}

func readfiles() {
	var result []string
	var dmipath string

	errs := filepath.Walk(AppConfig.IconsFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			dirindex := strings.Index(path, "icons")
			if dirindex < 0 {
				dmipath = "not found"
			} else {
				dmipath = path[dirindex:]
			}
			result = append(result, path)
			extensionless_name := info.Name()[:len(info.Name())-4]
			folder := strings.Replace(path, "/"+extensionless_name+".png", "", 1)
			i := strings.LastIndex(folder, "/") + 1
			folder = folder[i:]
			icon := Icon{extensionless_name, dmipath, path, folder}

			icons[extensionless_name] = icon
			orderbydmi(folder, &icon)
			return nil
		})

	if errs != nil {
		log.Println(errs)
	}
}

func main() {
	icons = make(map[string]Icon)
	icons_ordered_by_dmi = make(map[string][]*Icon)
	readconfig()
	readfiles()
	log.Println("ready")

	router := gin.Default()
	router.GET("/dmis", getfiles)
	//router.GET("/dmi/:filename", geticon)
	router.GET("/icon/search/:search", searcher)
	router.GET("/dmi/search/:search", searcherByDmi)

	scope := os.Getenv("SCOPE")
	if scope == "PROD" {
		log.Fatal(autotls.Run(router, "vgutils.com.ar"))
		return
	}
	router.Run("0.0.0.0:17011")
}

func readconfig() {
	f, err := os.Open("conf.yml")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&AppConfig)

	if err != nil {
		log.Fatalln(err)
	}
}

func searcher(c *gin.Context) {
	middleend(c)
	search := c.Param("search")
	var results []Icon
	for key, val := range icons {
		if strings.Contains(key, search) {
			results = append(results, val)
		}
	}
	c.JSON(http.StatusOK, results)
}

func searcherByDmi(c *gin.Context) {
	middleend(c)
	dmi := c.Param("search")

	r := regexp.MustCompile(`^[0-9a-zA-Z_-]+$`)
	if !r.MatchString(dmi) {
		c.JSON(http.StatusBadRequest, "what the dog doin")
		return
	}

	var result []Icon
	if icons, found := icons_ordered_by_dmi[dmi]; found {
		for _, f := range icons {
			result = append(result, *f)
		}
	} else {
		c.JSON(http.StatusNoContent, "bad luck")
		return
	}

	c.JSON(http.StatusOK, result)
}

func getfiles(c *gin.Context) {
	middleend(c)
	c.JSON(http.StatusOK, icons)
}

func geticon(c *gin.Context) {
	middleend(c)
	filename := c.Param("filename")
	if _, ok := icons[filename]; !ok {
		c.JSON(http.StatusNoContent, "no")
		return
	}
	filepath := icons[filename].filepath
	c.Writer.Header().Set("Content-Type", "image/png")
	c.File(filepath)
}

func middleend(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "public")
	c.Writer.Header().Set("Cache-Control", "max-age=86400") // a day
	c.Writer.Header().Set("Access-Control-Allow-Origin", "https://zth--.github.io")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
}
