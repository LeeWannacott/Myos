package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	parent_folder_path := flag.String("p", "", "folder path in current directory with sub folder containing folders with sprites; example: 'animations_folder'")
	help := flag.Bool("h", false, "Display help")

	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	var wg sync.WaitGroup
	if parent_folder_path != nil {
		parent_folder, err := os.ReadDir(filepath.Join(pwd, *parent_folder_path))
		if err != nil {
			log.Fatal(err)
		}
		for i, sub_folder := range parent_folder {
			if err != nil {
				fmt.Println(err)
			}
			sub_folder_path := filepath.Join(pwd, *parent_folder_path, sub_folder.Name())

			amount_of_sprites, folder_names, sprite_height, sprite_width := iterate_folder(sub_folder_path, i)
			if len(amount_of_sprites) == len(folder_names) {
				for i, folder_name := range folder_names {
					wg.Add(1)
					go func(i int, folder_name string) {
						defer wg.Done()
						make_spritesheet(i, folder_name, sub_folder_path, sprite_height, sprite_width, amount_of_sprites)
					}(i, folder_name)
				}
				wg.Wait()
			}
		}
	}
	fmt.Println(time.Since(start))
}

func make_spritesheet(i int, folder_name string, sub_folder_path string, sprite_height int, sprite_width int, amount_of_sprites []int) {
	spritesheet_width := 8
	background_type := "transparent"
	geometry_size := fmt.Sprintf("%vx%v", sprite_height, sprite_width)
	filter_type := "Catrom"
	spritesheet_height := math.Ceil(float64(amount_of_sprites[i]/spritesheet_width) + 1)
	input_folder_path := filepath.Join((sub_folder_path), folder_name, "/*")
	tile_size := fmt.Sprintf("%vx%v", spritesheet_width, spritesheet_height)
	sprite_name := fmt.Sprintf("%s_f%d_v%v.png", folder_name, amount_of_sprites[i], spritesheet_height)
	out, err := exec.Command("montage", input_folder_path, "-geometry", geometry_size, "-tile", tile_size,
		"-background", background_type, "-filter", filter_type, sprite_name).CombinedOutput()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	fmt.Println("Output: ", string(out), sprite_name)
}

func iterate_folder(file_path_to_walk string, index int) ([]int, []string, int, int) {
	is_first_sprite_in_directory := true
	folder_names := []string{}
	amount_of_sprites := []int{}
	sprite_height := 0
	sprite_width := 0

	is_containing_folder := true
	filepath.Walk(file_path_to_walk, func(path string, info os.FileInfo, err error) error {
		if !is_containing_folder {
			if err != nil {
				log.Fatalf(err.Error())
			}
			if info.IsDir() {
				folder_path, err := os.ReadDir(path)
				if err != nil {
					log.Fatalf(err.Error())
				}
				amount_of_sprites = append(amount_of_sprites, len(folder_path))
				folder_names = append(folder_names, info.Name())
			}
			if !info.IsDir() && is_first_sprite_in_directory {
				if reader, err := os.Open(path); err == nil {
					defer reader.Close()
					m, _, err := image.Decode(reader)
					if err != nil {
						log.Fatal(err)
					}
					bounds := m.Bounds()
					w := bounds.Dx()
					h := bounds.Dy()
					sprite_height = h
					sprite_width = w
					is_first_sprite_in_directory = false
				}
			}
		}
		is_containing_folder = false
		return nil
	})
	return amount_of_sprites, folder_names, sprite_height, sprite_width
}
