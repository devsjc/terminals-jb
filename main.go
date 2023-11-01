package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"text/template"
)

// --- Terminal color theme template files ---

//go:embed templates/iterm2.xml
var iterm2TemplateEmbed string
//go:embed templates/kitty.conf
var kittyTemplateEmbed string

type terminalData struct {
    Extension string
    Template string
}

var TERMINAL_TEMPLATE_MAP = map[string]terminalData{
        "iterm2": terminalData{
            Extension: ".itermcolors", 
            Template: iterm2TemplateEmbed,
        },
        "kitty": terminalData{
            Extension: ".conf",
            Template: kittyTemplateEmbed,
        },
    }

// --- Palette structs ---

// ColorRepr defines the different representations of a color
type ColorRepr struct {
    Hex string `json:"hex"`
    RGB []int `json:"rgb"`
    Xterm int `json:"xterm"`
}

// Palette is a map of color names to a dict of color representations
type Palette map[string]ColorRepr

// --- Ansi Map structs ---

// AnsiMapItem defines the corrolation beween an Ansi color code and a 
// color name from the palette
type AnsiMapItem struct {
    Key string `json:"key"`
    Col string `json:"col"`
}

// TemplateData defines all the required data for rendering a template
type TemplateData struct {
    Palette Palette
    StyleName string
    AnsiMap []AnsiMapItem
}

func main() {

    // Fetch palettes from vim-jb repository
    resp, err := http.Get("https://raw.githubusercontent.com/devsjc/vim-jb/main/autoload/palettes.json")
    if err != nil {
        panic(err)
    }
    // Parse palettes json into a go struct
    var palettes map[string]Palette
    err = json.NewDecoder(resp.Body).Decode(&palettes)
    if err != nil {
        panic(err)
    }

    // Read ansi color map from file
    ansiJson, err := os.ReadFile("ansipalettemap.json")
    if err != nil {
        panic(err)
    }
    // Parse color map json into a go struct
    var ansiMap map[string][]AnsiMapItem
    err = json.Unmarshal(ansiJson, &ansiMap)
    if err != nil {
        panic(err)
    }


    // Render templates
    for stylename, palette := range palettes {

        for terminalName, terminalData := range TERMINAL_TEMPLATE_MAP {

            // Create template data
            data := TemplateData{
                Palette: palette,
                StyleName: stylename,
                AnsiMap: ansiMap[terminalName],
            }

            // Create template
            t, err := template.New("iterm2").Funcs(
                template.FuncMap{
                    "div": func(a, b int) float32 { return float32(a) / float32(b) },
                },
            ).Parse(terminalData.Template)
            if err != nil {
                panic(err)
            }

            // Create the output file
            f, err := os.Create("colors/jb-" + stylename + terminalData.Extension)
            if err != nil {
                panic(err)
            }
            defer f.Close()
            
            // Render the template into the output file
            err = t.Execute(f, data)
            if err != nil {
                panic(err)
            }

        }
    }
}
