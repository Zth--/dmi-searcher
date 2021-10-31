package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
)

type Dmi struct {
	Name string `json:"dmi"`
	Url  string `json:"url"`
	Icon []Icon `json:"icons"`
}

type Icon struct {
	Filepath string `json:"filepath"`
	Name     string `json:"name"`
	Dmi      string `json:"dmi"`
}

var icons map[string]Icon

func readfiles() {
	var result []string
	var dmipath string

	vgicon_dir := "https://github.com/vgstation-coders/vgstation13/tree/Bleeding-Edge/"

	errs := filepath.Walk("../icons",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirindex := strings.Index(path, "icons")
				if dirindex < 0 {
					dmipath = "not found"
				} else {
					dmipath = vgicon_dir + path[dirindex:] + ".dmi"
				}
				return nil
			}
			result = append(result, path)
			extensionless_name := info.Name()[:len(info.Name())-4]
			icons[extensionless_name] = Icon{path, extensionless_name, dmipath}
			log.Println(extensionless_name, path)
			return nil
		})

	if errs != nil {
		log.Println(errs)
	}
}

func main() {
	icons = make(map[string]Icon)
	readfiles()
	log.Println("ready")
	router := gin.Default()
	router.GET("/dmis", getfiles)
	router.GET("/dmi/:filename", geticon)
	router.GET("/dmi/search/:search", searcher)

	router.Run("0.0.0.0:17011")
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

func getfiles(c *gin.Context) {
	middleend(c)
	c.JSON(http.StatusOK, icons)
}

func geticon(c *gin.Context) {
	middleend(c)
	filename := c.Param("filename")
	filepath := icons[filename].Filepath
	log.Println(filepath)
	c.Writer.Header().Set("Content-Type", "image/png")
	c.File(filepath)
}

func middleend(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
}
