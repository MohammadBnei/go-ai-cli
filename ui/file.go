package ui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disiqueira/gotree"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
)

func FileSelectionFzf(path string) (fileContents []string, err error) {
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return
		}
	}

	var selected []os.FileInfo

FileLoop:
	for {
		files, _err := os.ReadDir(path)
		if _err != nil {
			fmt.Println("Error while getting current working directory:", _err)
			err = errors.Join(errors.New("error while getting current working directory : "), _err)
			return
		}
		files = append(files, &myFileInfo{"..", 0, 0, time.Now(), true})
		files = lo.Filter(files, func(f os.DirEntry, _ int) bool {
			if f.IsDir() {
				return true
			}
			fileContent, err := os.ReadFile(path + "/" + f.Name())
			if err != nil {
				return false
			}
			return strings.Contains(http.DetectContentType(fileContent), "text/plain") || strings.Contains(f.Name(), ".svelte")
		})

		idx, _err := fuzzyfinder.FindMulti(
			files,
			func(i int) string {
				return files[i].Name()
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				if files[i].IsDir() {
					root := gotree.New(files[i].Name())
					subFiles, _err := os.ReadDir(path + "/" + files[i].Name())
					if _err != nil {
						return "📁"
					}
					for _, f := range subFiles {
						sub := root.Add(f.Name())
						if f.IsDir() {
							subFiles, err := os.ReadDir(path + "/" + files[i].Name())
							if err == nil {
								for _, f := range subFiles {
									sub.Add(f.Name())
								}
							}
						}
					}

					return root.Print()
				}
				fileContent, err := os.ReadFile(path + "/" + files[i].Name())
				if err != nil {
					return fmt.Sprintf("Error while reading file: %s\n", err)
				}
				return fmt.Sprintf("File: %s\nType: %s\nLength: %d\nContent: %s\n",
					files[i].Name(),
					http.DetectContentType(fileContent),
					len(string(fileContent)),
					AddReturnOnWidth(w/3-1, string(fileContent)),
				)
			}))

		if _err != nil {
			err = _err
			return
		}
		if len(idx) == 1 {
			file := files[idx[0]]

			switch {
			case file.Name() == "..":
				path = filepath.Dir(path)
			case file.IsDir():
				path += "/" + file.Name()
			default:
				selected = lo.Map[int, os.FileInfo](idx, func(i int, _ int) os.FileInfo {
					info, err := files[i].Info()
					if err != nil {
						return nil
					}
					return info
				})
				break FileLoop
			}
		} else {
			selected = lo.Map[int, os.FileInfo](idx, func(i int, _ int) os.FileInfo {
				info, err := files[i].Info()
				if err != nil {
					return nil
				}
				return info
			})
			break FileLoop
		}
	}

	for _, file := range selected {
		if file.IsDir() {
			fmt.Printf("%s is a directory, not adding it.\n", file.Name())
			continue
		}

		fileContent, _err := os.ReadFile(path + "/" + file.Name())
		if _err != nil {
			err = _err
			return
		}

		fileContents = append(fileContents, fmt.Sprintf("// Filename : %s\n%s", file.Name(), fileContent))

		fmt.Println("loaded file:", file.Name())
	}

	return
}
func PathSelectionFzf(startPath string) (path string, err error) {
	if startPath == "" {
		startPath, err = os.Getwd()
		if err != nil {
			return
		}
	}

	path = startPath

FileLoop:
	for {
		dirs, _err := os.ReadDir(path)
		if _err != nil {
			err = errors.Join(errors.New("error while getting current working directory : "), _err)
			return
		}
		dirs = lo.Filter[os.DirEntry](dirs, func(item os.DirEntry, _ int) bool {
			return item.IsDir()
		})
		dirs = append(dirs, &myFileInfo{".", 0, 0, time.Now(), true}, &myFileInfo{"..", 0, 0, time.Now(), true})

		id, _err := fuzzyfinder.Find(
			dirs,
			func(i int) string {
				return dirs[i].Name()
			},
		)
		if _err != nil {
			err = _err
			return
		}

		dir := dirs[id]
		switch dir.Name() {
		case "..":
			path = filepath.Dir(path)
		case ".":
			return
		default:
			path += "/" + dir.Name()
			break FileLoop
		}
	}

	return
}

func AddToFile(content []byte, filename string, withTimestamp bool) error {
	if filename == "" {
		filePrompt := promptui.Prompt{
			Label: "specify a filename (with extension)",
		}
		var err error
		filename, err = filePrompt.Run()
		if err != nil {
			return err
		}
	}

	if strings.Contains(filename, "/") {
		splitted := strings.Split(filename, "/")
		dw := strings.Join(splitted[:len(splitted)-1], "/")

		if _, err := os.Stat(dw); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dw, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created directory : " + dw)
		}
	}

	if withTimestamp {
		content = append([]byte(time.Now().Format("15:04:05 --- \n")), content...)
	}

	if _, err := os.Stat(filename); err == nil {
		fileContent, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		content = append([]byte("\n\n"), content...)
		content = append(fileContent, content...)
	}

	os.WriteFile(filename, content, os.ModePerm)

	fmt.Println("saved to", filename)

	return nil

}

func SaveToFile(content []byte, filename string) error {
	if filename == "" {
		var err error
		filename, err = StringPrompt("specify a filename (with extension)")
		if err != nil {
			return err
		}
	}

	if strings.Contains(filename, "/") {
		splitted := strings.Split(filename, "/")
		dw := strings.Join(splitted[:len(splitted)-1], "/")

		if _, err := os.Stat(dw); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dw, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created directory : " + dw)
		}
	}

	if _, err := os.Stat(filename); err == nil {
		replaceSelect := promptui.Select{
			Label: filename + " exists. Replace ?",
			Items: []string{"Yes", "No"},
		}

		i, _, err := replaceSelect.Run()
		if err != nil {
			return err
		}

		if i == 1 {
			return nil
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(content)
	fmt.Println("saved to", filename)

	return nil
}

func AddReturnOnWidth(w int, str string) string {
	splitted := strings.Split(str, " ")
	// acc := len(splitted[0])
	// for i := 1; i < len(splitted); i++ {
	// 	acc += len(splitted[i])
	// 	if acc > w {
	// 		splitted[i-1] += "\n"
	// 		acc = 0
	// 	}
	// }
	// str = strings.Join(splitted, " ")

	characterCount := 0

	return lo.Reduce[string, string](splitted, func(acc string, elem string, id int) string {
		characterCount += len(" " + elem)
		if characterCount > w {
			acc += "\n" + elem
			characterCount = 0
			return acc
		}
		if id != 0 {
			acc += " "
		}
		acc += elem

		return acc
	}, "")
	// return str
}
