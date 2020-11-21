// The following file is adapted from https://github.com/syohex/gowsay,
// with changes that were necessary for the satstack project. The
// original author was not explicit with his copyright notice, so
// we're attributing the contents of this file to him.
//
// Copyright (C) 2016 Shohei YOSHIDA
// Copyright (C) 2020 Anirudha Bose
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package fortunes

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"io"
	"log"
	"text/template"

	"github.com/mattn/go-runewidth"
	"github.com/mitchellh/go-wordwrap"
)

type face struct {
	Eyes     string
	Tongue   string
	Thoughts string
}

var (
	cow = `        {{.Thoughts}}   ^__^
         {{.Thoughts}}  ({{.Eyes}})\_______
            (__)\       )\/\
             {{.Tongue}} ||----w |
                ||     ||
`
	columns = 40
)

func newFace() *face {
	return &face{
		Eyes:   "₿₿",
		Tongue: "  ",
	}
}

func readInput(msg string) []string {
	var msgs []string

	expand := strings.Replace(msg, "\t", "        ", -1)

	tmp := wordwrap.WrapString(expand, uint(columns))
	for _, s := range strings.Split(tmp, "\n") {
		msgs = append(msgs, s)
	}

	return msgs
}

func setPadding(msgs []string, width int) []string {
	var ret []string
	for _, m := range msgs {
		s := m + strings.Repeat(" ", width-runewidth.StringWidth(m))
		ret = append(ret, s)
	}

	return ret
}

func constructBallon(f *face, msgs []string, width int) string {
	var borders []string
	line := len(msgs)

	f.Thoughts = "\\"
	if line == 1 {
		borders = []string{"<", ">"}
	} else {
		borders = []string{"/", "\\", "\\", "/", "|", "|"}
	}

	var lines []string

	topBorder := " " + strings.Repeat("_", width+2)
	bottomBoder := " " + strings.Repeat("-", width+2)

	lines = append(lines, topBorder)
	if line == 1 {
		s := fmt.Sprintf("%s %s %s", borders[0], msgs[0], borders[1])
		lines = append(lines, s)
	} else {
		s := fmt.Sprintf(`%s %s %s`, borders[0], msgs[0], borders[1])
		lines = append(lines, s)
		i := 1
		for ; i < line-1; i++ {
			s = fmt.Sprintf(`%s %s %s`, borders[4], msgs[i], borders[5])
			lines = append(lines, s)
		}
		s = fmt.Sprintf(`%s %s %s`, borders[2], msgs[i], borders[3])
		lines = append(lines, s)
	}

	lines = append(lines, bottomBoder)
	return strings.Join(lines, "\n")
}

func maxWidth(msgs []string) int {
	max := -1
	for _, m := range msgs {
		l := runewidth.StringWidth(m)
		if l > max {
			max = l
		}
	}

	return max
}

func renderCow(f *face, w io.Writer) {
	t := template.Must(template.New("cow").Parse(cow))

	if err := t.Execute(w, f); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func Fortune() {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(bitcoinFortunes)

	inputs := readInput(bitcoinFortunes[n])
	width := maxWidth(inputs)
	messages := setPadding(inputs, width)

	f := newFace()
	balloon := constructBallon(f, messages, width)

	fmt.Println(balloon)
	renderCow(f, os.Stdout)
}
