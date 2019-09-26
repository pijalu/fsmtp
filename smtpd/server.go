/*Package smtpd Smtpd server implementation

Copyright © 2019 Pierre Poissinger
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package smtpd

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
)

// Login handles a login command with username and password.
func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return &Session{
		config:   bkd.config,
		username: username,
		password: password,
	}, nil
}

// AnonymousLogin a logoi
func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{
		config:   bkd.config,
		username: anonymous,
	}, nil
	//return nil, smtp.ErrAuthRequired
}

//Mail Process the from
func (s *Session) Mail(from string) error {
	s.from = from
	return nil
}

//Rcpt Process the To
func (s *Session) Rcpt(to string) error {
	s.to = to
	return nil
}

//Data Process the data
func (s *Session) Data(r io.Reader) error {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
		return err
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
		return err
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		log.Printf("Ignoring non-multipart mail !")
		return nil
	}

	// Create a unique prefix
	uuidPart, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
		return err
	}
	uuidPrefix := uuidPart.String()
	// Create a path prefix
	pathPrefix := path.Join(
		".",
		s.config.OutputPath,
		uuidPrefix[0:2],
		uuidPrefix,
	)

	// Parse multipart
	mr := multipart.NewReader(msg.Body, params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return nil
		}
		// Error
		if err != nil {
			log.Fatal(err)
			return err
		}

		// Only save files
		if p.FileName() == "" {
			//log.Printf("Ignoring non-file mail !")
			continue
		}

		// Prep path
		if err := os.MkdirAll(pathPrefix, 0755); err != nil {
			log.Fatal(err)
			return err
		}

		// Create decoder
		var r io.Reader
		encoding := p.Header.Get("Content-Transfer-Encoding")
		switch encoding {
		case "base64":
			r = base64.NewDecoder(base64.StdEncoding, p)
			break
		case "base32":
			r = base32.NewDecoder(base32.StdEncoding, p)
			break
		default:
			err := fmt.Errorf("Unknow encoding: %s", encoding)
			log.Fatal(err)
			return err
		}

		// Prep file
		targetName := path.Join(
			pathPrefix,
			p.FileName(),
		)

		// Use a temp file name
		tempTargetName := fmt.Sprintf("%s.tmp",
			targetName)

		// Create
		w, err := os.Create(tempTargetName)
		if err != nil {
			log.Fatal(err)
			return err
		}
		// Copy
		copied, err := io.Copy(w, r)
		// close now
		w.Close()
		if err != nil {
			log.Fatal(err)
			return err
		}

		// Rename
		os.Rename(tempTargetName, targetName)

		// Log
		log.Printf("Saved file %s [%d byte(s)]",
			targetName,
			copied)
	}
}

//Reset the session
func (s *Session) Reset() {}

//Logout is called on logout
func (s *Session) Logout() error {
	return nil
}
