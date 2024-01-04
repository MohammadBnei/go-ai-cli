package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/disiqueira/gotree"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
)

// func AskForEditImage(filePath string) error {
// 	imagePrompt := promptui.Prompt{
// 		Label: "what is the purpose",
// 	}

// 	desc, err := imagePrompt.Run()
// 	if err != nil {
// 		return err
// 	}

// 	sizePrompt := promptui.Select{
// 		Label: "Pick a size",
// 		Items: []string{openai.CreateImageSize256x256, openai.CreateImageSize512x512, openai.CreateImageSize1024x1024},
// 	}

// 	_, size, err := sizePrompt.Run()

// 	if err != nil {
// 		size = openai.CreateImageSize256x256
// 	}

// 	selectedImage := filePath
// 	if filePath != "" {

// 		useLastPrompt := promptui.Select{
// 			Label: fmt.Sprintf("Use last (%s)", filePath),
// 			Items: []string{"Yes", "no"},
// 		}

// 		_, useLast, err := useLastPrompt.Run()
// 		if err != nil {
// 			fmt.Println(err)
// 			useLast = "no"
// 		}

// 		if useLast == "no" {
// 			selectedImage = GetPngFilePath()
// 		}
// 	} else {
// 		selectedImage = GetPngFilePath()
// 	}

// 	if selectedImage == "" {
// 		return errors.New("no image selected")
// 	}

// 	b, err := service.EditImage(selectedImage, desc, size)
// 	if err != nil {
// 		return err
// 	}

// 	return SaveToFile(b, "")

// }

func fuzzHandleDir(file os.FileInfo, cwd string) string {
	root := gotree.New(file.Name())
	subFiles, err := ioutil.ReadDir(cwd + "/" + file.Name())
	if err != nil {
		return "📁"
	}
	subFiles = lo.Filter[os.FileInfo](subFiles, func(f os.FileInfo, _ int) bool {
		switch {
		case f.IsDir():
			return true
		case filepath.Ext(cwd+"/"+f.Name()) == ".png":
			return true
		}

		return false
	})
	for _, f := range subFiles {
		sub := root.Add(f.Name())
		if f.IsDir() {
			subFiles, err := ioutil.ReadDir(cwd + "/" + file.Name())
			subFiles = lo.Filter[os.FileInfo](subFiles, func(f os.FileInfo, _ int) bool {
				switch {
				case f.IsDir():
					return true
				case filepath.Ext(cwd+"/"+f.Name()) == ".png":
					return true
				}

				return false
			})
			if err == nil {
				for _, f := range subFiles {
					sub.Add(f.Name())
				}
			}
		}
	}

	return root.Print()
}

func GetPngFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for {
		files, err := ioutil.ReadDir(cwd)
		if err != nil {
			fmt.Println("Error while getting current working directory:", err)
			return ""
		}
		files = lo.Filter[os.FileInfo](files, func(f os.FileInfo, _ int) bool {
			switch {
			case f.IsDir():
				return true
			case filepath.Ext(cwd+"/"+f.Name()) == ".png":
				return true
			}

			return false
		})
		files = append(files, &myFileInfo{"..", 0, 0, time.Now(), true})

		idx, err := fuzzyfinder.Find(
			files,
			func(i int) string {
				return files[i].Name()
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				if files[i].IsDir() {
					return fuzzHandleDir(files[i], cwd)
				}

				return fmt.Sprintf("File: %s\nSize: %d",
					files[i].Name(),
					files[i].Size(),
				)
			}))

		if err != nil {
			fmt.Println(err)
			return ""
		}

		file := files[idx]

		switch {
		case file.Name() == "..":
			cwd = filepath.Dir(cwd)
		case file.IsDir():
			cwd += "/" + file.Name()
		default:
			return cwd + "/" + file.Name()
		}
	}
}
