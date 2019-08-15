package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
String Res Extractor ...
*/
func main() {

	//folderPath := "../resource_for_testing/"

	if len(os.Args) < 2 {
		fmt.Println("Please give me the project's folder path you want to check")
		return
	}

	folderPath := os.Args[1]
	fmt.Println("check", folderPath)

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		fmt.Println("<!-- ERROR -->: Folder was not found!")
	}

	result := true
	isStringFileExisted := false

	stringResTable := make(map[string]string)
	var layoutFiles []string
	var warningStrings []string
	stringResTableForNewCreated := make(map[string]string)

	var regForLayoutFile = regexp.MustCompile(`\/res\/layout\/.*\.xml`)

	//var isFirstChange = true
	var stringsXMLfilePath = ""

	err := filepath.Walk(folderPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if regForLayoutFile.MatchString(path) {
				fmt.Println("layout was found --> " + path)
				layoutFiles = append(layoutFiles, path)
			}

			if strings.HasSuffix(path, "values/strings.xml") {
				fmt.Println("file for string resource was found --> " + path)
				stringsXMLfilePath = path
				// @BH_Lin ---------------------------------------------------->
				// Read file line by line
				file, err := os.Open(path)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				isStringFileExisted = true

				//keymap := make(map[string]string)

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					lineText := strings.TrimSpace(scanner.Text())
					var re = regexp.MustCompile(`\<string name=\".*<\/string>`)

					isValidLine := re.MatchString(lineText)
					if isValidLine {
						fmt.Println(lineText)
						// <string name="signing_password_title">修改密碼</string>
						regCheckingString1_1 := regexp.MustCompile(`.*<string name=\"`)
						temp1_1 := regCheckingString1_1.ReplaceAllString(lineText, ``)
						//fmt.Println("step 1_1: " + temp1_1)
						regCheckingString1_2 := regexp.MustCompile(`\".*`)
						stringID := regCheckingString1_2.ReplaceAllString(temp1_1, ``)

						regCheckingStringForVariable := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z_$0-9]*$`)
						if !regCheckingStringForVariable.MatchString(stringID) {
							fmt.Println("FAIL> key=" + stringID)
							continue
						}

						fmt.Println("key=" + stringID)

						regCheckingString2_1 := regexp.MustCompile(`.*<string name=\".*\">`)
						temp2_1 := regCheckingString2_1.ReplaceAllString(lineText, ``)
						//fmt.Println("step 2_1: " + temp1_1)
						regCheckingString2_2 := regexp.MustCompile(`<\/string>.*`)
						stringValue := regCheckingString2_2.ReplaceAllString(temp2_1, ``)
						fmt.Println("value= " + stringValue)
						fmt.Println("---------------------------------------------------------------")

						stringResTable[stringID] = stringValue
					} else {
						fmt.Println("WARNING: " + lineText)
						warningStrings = append(warningStrings, lineText)
					}
				}

				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
				// @BH_Lin ----------------------------------------------------<
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	} else {
		count := 1
		for key, value := range stringResTable {
			fmt.Println("#" + strconv.Itoa(count) + " :StringID=___" + key + "___, Text=___" + value + "___")
			count++
		}
	}

	// Check Layout XML file
	for index, layoytFilePath := range layoutFiles {
		fmt.Println("@@@ " + strconv.Itoa(index) + " :Check file=___" + layoytFilePath)

		file, err := os.Open(layoytFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		var linesForNewLayout []string

		var isNeededToReplaceID = false

		for scanner.Scan() {
			//lineText := strings.TrimSpace(scanner.Text())
			lineText := scanner.Text()

			regCheckingStringForAndroidText := regexp.MustCompile(`.*android\:text=\".*\"`)

			isAndroidTextLine := regCheckingStringForAndroidText.MatchString(lineText)
			if !isAndroidTextLine {
				//fmt.Println("LINE: " + lineText)
				linesForNewLayout = append(linesForNewLayout, lineText)
				continue
			}

			regCheckingStringForValidAndroidText := regexp.MustCompile(`.*android\:text=\"\@string.*\"`)
			isValidAndroidTextLine := regCheckingStringForValidAndroidText.MatchString(lineText)

			if isValidAndroidTextLine {
				//fmt.Println("LINE: " + lineText)
				linesForNewLayout = append(linesForNewLayout, lineText)
				continue
			}

			// Try to extract the hardcodedtext
			//t.replace(/.*android\:text=\"/g, '').replace(/\".\/>/g, '')
			fmt.Println("Ready to Extract String ___" + lineText + "___")
			regCheckingStringForValidAndroidTextPrefix := regexp.MustCompile(`.*android\:text=\"`)
			tempLine := regCheckingStringForValidAndroidTextPrefix.ReplaceAllString(lineText, ``)
			regCheckingStringForValidAndroidTextSuffix1 := regexp.MustCompile(`\".\/>`)
			regCheckingStringForValidAndroidTextSuffix2 := regexp.MustCompile(`\"$`)
			if regCheckingStringForValidAndroidTextSuffix1.MatchString(tempLine) {
				tempLine = regCheckingStringForValidAndroidTextSuffix1.ReplaceAllString(tempLine, ``)
			} else if regCheckingStringForValidAndroidTextSuffix2.MatchString(tempLine) {
				tempLine = regCheckingStringForValidAndroidTextSuffix2.ReplaceAllString(tempLine, ``)
			} else {
				linesForNewLayout = append(linesForNewLayout, lineText)
				continue
			}

			if len(tempLine) == 0 && tempLine == "" {
				//fmt.Println("LINE: " + lineText)
				continue
			}

			var stringIDForReplacement = ""
			// Check if tmpeLine is existed
			for stringID, text := range stringResTable {
				if text == tempLine {
					stringIDForReplacement = stringID
					break
				}
			}

			// if there is no the same string text, try to create a new one
			if stringIDForReplacement == "" && len(stringIDForReplacement) == 0 {
				rand.Seed(time.Now().UTC().UnixNano())
				stringIDForReplacement = fmt.Sprintf("%d%s", time.Now().UnixNano(), randomString(10))

				stringIDForReplacement = "stringid_" + stringIDForReplacement
				stringResTableForNewCreated[stringIDForReplacement] = tempLine
				stringResTable[stringIDForReplacement] = tempLine
			}

			fmt.Println("Extracted String ==> ___" + tempLine + "___ for new Id: " + stringIDForReplacement)

			newLineText := strings.Replace(lineText, tempLine, "@string/"+stringIDForReplacement, 1)
			fmt.Println("OK to set id --> " + newLineText + " for text ___ " + tempLine + "___s")
			//fmt.Println("LINE: " + newLineText)
			linesForNewLayout = append(linesForNewLayout, newLineText)
			isNeededToReplaceID = true
		}

		// Rewrite File
		//if isNeededToReplaceID && isFirstChange {
		if isNeededToReplaceID {
			//isFirstChange = false
			file, err := os.OpenFile(layoytFilePath, os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				log.Fatalf("failed creating file: %s", err)
			}

			datawriter := bufio.NewWriter(file)

			for lineIndex, data := range linesForNewLayout {

				if lineIndex < len(linesForNewLayout)-1 {
					//fmt.Println("add new line")
					_, _ = datawriter.WriteString(data + "\n")
				} else {
					//fmt.Println("NOT add new line")
					_, _ = datawriter.WriteString(data)
				}
			}

			datawriter.Flush()
			file.Close()
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if stringsXMLfilePath != "" {

		file, err := os.OpenFile(stringsXMLfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		} else {
			stringXMLdatawriter := bufio.NewWriter(file)

			for stringID, data := range stringResTableForNewCreated {

				theNewStringResource := "<string name=\"" + stringID + "\">" + data + "</string>\n"
				_, _ = stringXMLdatawriter.WriteString(theNewStringResource)
			}
			stringXMLdatawriter.Flush()
			file.Close()
		}

	}

	if !isStringFileExisted {
		fmt.Println("\nX> There is no string.xml file.")
	}
	if result == true {
		fmt.Println("\n^_^b OK> All")
	} else {
		fmt.Println("\nX_X! NG>")
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return strings.ToLower(string(bytes))
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
