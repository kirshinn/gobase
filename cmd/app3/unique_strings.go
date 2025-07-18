package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func unique(input io.Reader, output io.Writer) error {
	in := bufio.NewScanner(input)
	//var prev string

	alreadySeen := make(map[string]bool)

	for in.Scan() {
		// преобразуем считанную строку в текст
		txt := in.Text()

		// проверка имеется ли в карте значение
		if _, found := alreadySeen[txt]; found {
			continue
		}

		//if txt == prev {
		//	continue
		//}

		//if txt < prev {
		//panic("file not sorted")
		//	return fmt.Errorf("file not sorted")
		//}

		//prev = txt

		alreadySeen[txt] = true

		_, err := fmt.Fprintln(output, txt)
		if err != nil {
			return err
		}

	}
	return nil
}

func main() {
	err := unique(os.Stdin, os.Stdout)
	if err != nil {
		panic(err.Error())
	}
}
