package main

import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	// "github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"log"
	// "encoding/json"
	//"encoding/hex"
	// "os"
	"github.com/fogleman/gg"
	"image/color"
)

var (
	Black   = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	White   = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	Pink    = color.RGBA{0xF2, 0x07, 0x65, 0xFF}
	Purple  = color.RGBA{0x85, 0x08, 0x41, 0xFF}
	Orange  = color.RGBA{0xFF, 0x66, 0x00, 0xFF}
	Blue0   = color.RGBA{0x1e, 0x8b, 0xb8, 0xff}
	Blue1   = color.RGBA{0x17, 0x6c, 0x8f, 0xff}
	Blue2   = color.RGBA{0x14, 0x5d, 0x7b, 0xff}
	Blue3   = color.RGBA{0x11, 0x4d, 0x66, 0xff}
	Blue4   = color.RGBA{0x0d, 0x3e, 0x52, 0xff}
	Blue5   = color.RGBA{0x0a, 0x2e, 0x3d, 0xff}
	Blue6   = color.RGBA{0x06, 0x1f, 0x29, 0xff}
	Blue7   = color.RGBA{0x03, 0x0f, 0x14, 0xff}
	Green1  = color.RGBA{0xbc, 0xcf, 0x48, 0xff}
	Green2  = color.RGBA{0xa7, 0xb8, 0x40, 0xff}
	Green3  = color.RGBA{0x92, 0xa1, 0x38, 0xff}
	Green4  = color.RGBA{0x7d, 0x8a, 0x30, 0xff}
	Green5  = color.RGBA{0x68, 0x73, 0x28, 0xff}
	Green6  = color.RGBA{0x53, 0x5c, 0x20, 0xff}
	Green7  = color.RGBA{0x3e, 0x45, 0x18, 0xff}
	Green8  = color.RGBA{0x29, 0x2e, 0x10, 0xff}
	Yellow1 = color.RGBA{0xf1, 0xea, 0x9b, 0xff}
	Yellow2 = color.RGBA{0xD6, 0xC5, 0x84, 0xff}
)


func main() {

	// tutoj otevře db

	db, err := sql.Open("sqlite3", "./lux.mbtiles")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	/*

	// tutoj vypíše metadata

	rows, err := db.Query("select name, value from metadata")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var v string
		err = rows.Scan(&k, &v)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(k, v)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	*/
	
	// tutoj získá dlaždicu

	// používají se TMS souřadnice
	// viz https://www.maptiler.com/how-to/coordinates-bounds-projection/
	tile := maptile.New(8471, 10805, 14)

	// ves Bofferdange (v Lucembursku) := maptile.New(8471, 10805, 14)

	fmt.Println(tile.Center())

	var data []byte
	err2 := db.QueryRow("SELECT tile_data FROM tiles WHERE zoom_level=? AND tile_row=? AND tile_column=?", tile.Z, tile.Y, tile.X).Scan(&data)
	if err2 != nil {
		log.Fatal(err2)
	}

	// OpenMapTiles je gzipped
	layers, err3 := mvt.UnmarshalGzipped(data)
	if err3 != nil {
		log.Fatal(err3)
	}


	// tutoj vykreslí dlaždicu

	resolution := 512.0

	dc := gg.NewContext(int(resolution), int(resolution))
	dc.SetRGB(255, 255, 255)
	dc.DrawRectangle(0, 0, resolution, resolution)
	dc.Fill()
	
	colors := make(map[string]color.Color)
	colors["water"] = Blue0
	colors["waterway"] = Blue1
	colors["transportation"] = Pink
	colors["building"] = Orange
	colors["park"] = Green1
	colors["landuse"] = Green3
	colors["landcover"] = White

	// souřadnice pro obrázek se počítají stylem 
	// dest = src / extent * resolution
	//
	// v dlaždici jsou uloženy i objekty mimo ní

	resolution = float64(resolution)
	
	for _, layer := range layers {

		tocolor, selected := colors[layer.Name]
		if !selected {
			// výchozí barva
			tocolor = Purple
		}
		dc.SetColor(tocolor)
		
		extent := float64(layer.Extent)

		// fc := geojson.NewFeatureCollection()
		
		for _, feat := range layer.Features {
			// fc.Append(feat)

			if geom, ok := feat.Geometry.(orb.Point); ok {
				// fmt.Println("pt", geom.X(), geom.Y())

				dc.DrawPoint(geom.X() / extent * resolution, geom.Y()  / extent * resolution, 1)
			}

			if geom, ok := feat.Geometry.(orb.LineString); ok {
				// fmt.Println(geom.X(), geom.Y())

				for _, pt := range geom {
					// fmt.Println("line", pt.X(), pt.Y())
					dc.LineTo(pt.X()  / extent * resolution , pt.Y() / extent * resolution)
				}
				dc.Stroke()				
			}

			if geom, ok := feat.Geometry.(orb.Polygon); ok {
				// fmt.Println(geom.X(), geom.Y())

				for _, ring := range geom {
					for _, pt := range ring {
						// fmt.Println("line", pt.X(), pt.Y())
						dc.LineTo(pt.X()  / extent * resolution , pt.Y() / extent * resolution)
					}
					dc.NewSubPath()
				}
				// dc.Stroke()				
				
				// Fill zatím ne
				dc.Fill()				
			}

		}

		// blob, _ := json.MarshalIndent(fc, "", "  ")

		fmt.Println(layer.Name)
		// os.Stdout.Write(blob)
	}	

	dc.SavePNG("out.png")
	fmt.Println("OK")
}