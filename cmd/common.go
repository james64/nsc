/*
 * Copyright 2018 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nats-io/nkeys"
	"github.com/spf13/cobra"
)

func OkToWrite(fp string) bool {
	// stdout
	if fp == "--" {
		return true
	}

	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return true
	}
	return false
}

func IsReadableFile(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil
}

func GetOutput(fp string) (*os.File, error) {
	var err error
	var f *os.File

	if fp == "--" {
		f = os.Stdout
	} else {
		_, err = os.Stat(fp)
		if err == nil {
			return nil, fmt.Errorf("%q already exists", fp)
		}
		if !os.IsNotExist(err) {
			return nil, err
		}

		f, err = os.Create(fp)
		if err != nil {
			return nil, fmt.Errorf("error creating output file %q: %v", fp, err)
		}
	}
	return f, nil
}

func IsStdOut(fp string) bool {
	return fp == "--"
}

func Write(fp string, data []byte) error {
	var err error
	var f *os.File

	f, err = GetOutput(fp)
	if err != nil {
		return err
	}
	if !IsStdOut(fp) {
		defer f.Close()
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing %q: %v", fp, err)
	}

	if !IsStdOut(fp) {
		if err := f.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func FormatKeys(keyType string, publicKey string, privateKey string) []byte {
	w := bytes.NewBuffer(nil)
	label := strings.ToUpper(keyType)

	if privateKey != "" {
		fmt.Fprintln(w, "************************* IMPORTANT *************************")
		fmt.Fprintln(w, "Your options generated NKEYs which can be used to create")
		fmt.Fprintln(w, "entities or prove identity.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Generated keys printed below are sensitive and should be")
		fmt.Fprintln(w, "treated as secrets to prevent unauthorized access.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "The private key is not saved by the tool. Please save")
		fmt.Fprintln(w, "it now as it will be required by the user to connect to NATS.")
		fmt.Fprintln(w, "The public key is saved and uniquely identifies the user.")
		fmt.Fprintln(w)

		fmt.Fprintf(w, "-----BEGIN %s PRIVATE KEY-----\n", label)
		fmt.Fprintln(w, privateKey)
		fmt.Fprintf(w, "------END %s PRIVATE KEY------\n", label)
		fmt.Fprintln(w)

		fmt.Fprintln(w, "*************************************************************")
		fmt.Fprintln(w)
	}

	if publicKey != "" {
		fmt.Fprintf(w, "-----BEGIN %s PUB KEY-----\n", label)
		fmt.Fprintln(w, publicKey)
		fmt.Fprintf(w, "------END %s PUB KEY------\n", label)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)

	return w.Bytes()
}

func ExtractToken(s string) string {
	// remove all the spaces
	re := regexp.MustCompile(`\s+`)
	w := re.ReplaceAllString(s, "")
	// remove multiple dashes
	re = regexp.MustCompile(`\-+`)
	w = re.ReplaceAllString(w, "-")

	// the token can now look like
	// -BEGINXXXXPUBKEY-token-ENDXXXXPUBKEY-
	re = regexp.MustCompile(`(?m)(\-BEGIN.+(JWT|KEY)\-)(?P<token>.+)(\-END.+(JWT|KEY)\-)`)
	// find the index of the token
	m := re.FindStringSubmatch(w)
	if len(m) > 0 {
		for i, name := range re.SubexpNames() {
			if name == "token" {
				return m[i]
			}
		}
	}
	return s
}

func FormatJwt(jwtType string, jwt string) []byte {
	w := bytes.NewBuffer(nil)

	label := strings.ToUpper(jwtType)
	fmt.Fprintf(w, "-----BEGIN %s JWT-----\n", label)
	fmt.Fprintln(w, jwt)
	fmt.Fprintf(w, "------END %s JWT------\n", label)
	fmt.Fprintln(w)

	return w.Bytes()
}

// parse expiration argument - supported are YYYY-MM-DD for absolute, and relative
// (m)inute, (h)our, (d)ay, (w)week, (M)onth, (y)ear expressions
func ParseExpiry(s string) (int64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}
	re := regexp.MustCompile(`(\d){4}-(\d){2}-(\d){2}`)
	if re.MatchString(s) {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return 0, err
		}
		return t.Unix(), nil
	}
	re = regexp.MustCompile(`(?P<count>\d+)(?P<qualifier>[mhdMyw])`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, err
		}
		count := int(v)
		if count == 0 {
			return 0, nil
		}
		dur := time.Duration(count)
		now := time.Now()
		switch m[2] {
		case "m":
			return now.Add(dur * time.Minute).Unix(), nil
		case "h":
			return now.Add(dur * time.Hour).Unix(), nil
		case "d":
			return now.AddDate(0, 0, count).Unix(), nil
		case "w":
			return now.AddDate(0, 0, 7*count).Unix(), nil
		case "M":
			return now.AddDate(0, count, 0).Unix(), nil
		case "y":
			return now.AddDate(count, 0, 0).Unix(), nil
		default:
			return 0, fmt.Errorf("unknown interval %q in %q", m[2], m[0])
		}
	}
	return 0, fmt.Errorf("couldn't parse expiry: %v", s)
}

func ParseNumber(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.ToUpper(s)
	re := regexp.MustCompile(`(\d+$)`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[0], 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	re = regexp.MustCompile(`(\d+)([K|M|G])`)
	m = re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, err
		}
		if m[2] == "K" {
			return v * 1000, nil
		}
		if m[2] == "M" {
			return v * 1000000, nil
		}
		if m[2] == "G" {
			return v * 1000000000, nil
		}
	}
	return 0, fmt.Errorf("couldn't parse number: %v", s)
}

func ParseDataSize(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.ToUpper(s)
	re := regexp.MustCompile(`(\d+$)`)
	m := re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[0], 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	re = regexp.MustCompile(`(\d+)([B|K|M])`)
	m = re.FindStringSubmatch(s)
	if m != nil {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, err
		}
		if m[2] == "B" {
			return v, nil
		}
		if m[2] == "K" {
			return v * 1000, nil
		}
		if m[2] == "M" {
			return v * 1000000, nil
		}
	}
	return 0, fmt.Errorf("couldn't parse data size: %v", s)
}

func list(cmd *cobra.Command, subDir string, extension string, label string) {
	s, err := getStore()
	if err != nil {
		cmd.Printf("error loading store: %v\n", err)
		return
	}

	a, err := s.List(subDir, extension)
	if err != nil {
		cmd.Printf("error listing %s: %v\n", label, err)
		return
	}

	if len(a) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 4, 8, 3, ' ', 0)
		for i, n := range a {
			fmt.Fprintf(w, "%d\t %s\n", i+1, n)
		}
		w.Flush()
	} else {
		cmd.Printf("No %s found\n", label)
	}
}

func LooksLikeNKey(s string, prefix byte) bool {
	pre := string(prefix)
	if len(s) == 109 || len(s) == 58 {
		pre = "S" + string(prefix)
		return strings.HasPrefix(s, pre) && !strings.Contains(s, string(filepath.Separator))
	}
	if len(s) == 56 {
		return strings.HasPrefix(s, pre) && !strings.Contains(s, string(filepath.Separator))
	}
	return false
}

func ParseNKey(s string) (nkeys.KeyPair, error) {
	if s[0] == 'S' {
		return nkeys.FromSeed([]byte(s))
	} else {
		return nkeys.FromPublicKey([]byte(s))
	}
}

func DefaultName(s string) string {
	if s == "" {
		return "(not specified)"
	}
	return s
}