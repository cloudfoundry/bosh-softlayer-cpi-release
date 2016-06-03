package util_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	bscutil "github.com/cloudfoundry/bosh-softlayer-cpi/util"
	"github.com/pkg/sftp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SshClient", func() {
	var (
		sshClient     bscutil.SshClient
		listener      net.Listener
		serverAddress string

		serverConfig *ssh.ServerConfig
		sshServer    *server
	)

	BeforeEach(func() {
		sshClient = bscutil.GetSshClient()

		serverHostKey, err := ssh.ParsePrivateKey([]byte(hostKey))
		Expect(err).NotTo(HaveOccurred())

		listener, serverAddress = newListener()

		serverConfig = &ssh.ServerConfig{}
		serverConfig.AddHostKey(serverHostKey)
		serverConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if conn.User() == "testuser" && string(password) == "testpass" {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.New("authentication failed")
		}

		sshServer = &server{
			mutex:        &sync.Mutex{},
			listener:     listener,
			serverConfig: serverConfig,
		}
	})

	JustBeforeEach(func() {
		go sshServer.Serve()
	})

	AfterEach(func() {
		sshServer.Shutdown()
	})

	Describe("ExecCommand", func() {
		BeforeEach(func() {
			sshServer.sessionChannelHandler = func(newChannel ssh.NewChannel) {
				defer GinkgoRecover()

				channel, requests, err := newChannel.Accept()
				Expect(err).NotTo(HaveOccurred())
				defer channel.Close()

				for req := range requests {
					switch req.Type {
					case "exec":
						var execMsg struct{ Command string }
						err := ssh.Unmarshal(req.Payload, &execMsg)
						Expect(err).NotTo(HaveOccurred())

						if req.WantReply {
							req.Reply(true, nil)
						}

						var exitStatusMsg struct{ Status uint32 }
						io.Copy(channel, strings.NewReader(execMsg.Command))
						channel.SendRequest("exit-status", false, ssh.Marshal(exitStatusMsg))
						return
					default:
						if req.WantReply {
							req.Reply(false, nil)
						}
					}
				}
			}
		})

		It("executes the command over ssh", func() {
			result, err := sshClient.ExecCommand("testuser", "testpass", serverAddress, "this is my command string")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("this is my command string"))
		})

		Context("when authentication fails", func() {
			It("returns an error", func() {
				_, err := sshClient.ExecCommand("testuser", "broken", serverAddress, "this is my command string")
				Expect(err).To(MatchError(ContainSubstring("handshake failed")))
			})
		})
	})

	Describe("sftp operations", func() {
		var tempDir, localDir, remoteDir string

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "sftp")
			Expect(err).NotTo(HaveOccurred())

			localDir, err = ioutil.TempDir(tempDir, "local")
			Expect(err).NotTo(HaveOccurred())

			remoteDir, err = ioutil.TempDir(tempDir, "remote")
			Expect(err).NotTo(HaveOccurred())

			sshServer.sessionChannelHandler = func(newChannel ssh.NewChannel) {
				defer GinkgoRecover()

				channel, requests, err := newChannel.Accept()
				Expect(err).NotTo(HaveOccurred())
				defer channel.Close()

				for req := range requests {
					switch req.Type {
					case "subsystem":
						var subsysMsg struct{ Subsystem string }
						err := ssh.Unmarshal(req.Payload, &subsysMsg)
						Expect(err).NotTo(HaveOccurred())
						Expect(subsysMsg.Subsystem).To(Equal("sftp"))

						sftpServer, err := sftp.NewServer(channel, channel, sftp.WithDebug(GinkgoWriter))
						Expect(err).NotTo(HaveOccurred())

						if req.WantReply {
							req.Reply(true, nil)
						}

						err = sftpServer.Serve()
						Expect(err).NotTo(HaveOccurred())
						return
					default:
						if req.WantReply {
							req.Reply(false, nil)
						}
					}
				}
			}
		})

		AfterEach(func() {
			os.RemoveAll(tempDir)
		})

		Describe("downloads", func() {
			var remoteFile string

			BeforeEach(func() {
				file, err := ioutil.TempFile(remoteDir, "sftp")
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				w := bufio.NewWriter(file)
				w.WriteString("file contents\non two lines")
				w.Flush()

				remoteFile = file.Name()
			})

			Describe("Download", func() {
				var destination *bytes.Buffer

				BeforeEach(func() {
					destination = &bytes.Buffer{}
				})

				It("writes the contents of the remote file to the writer", func() {
					err := sshClient.Download("testuser", "testpass", serverAddress, remoteFile, destination)
					Expect(err).NotTo(HaveOccurred())
					Expect(destination.Bytes()).To(Equal([]byte("file contents\non two lines")))
				})

				Context("when authentication fails", func() {
					It("returns an error", func() {
						err := sshClient.Download("testuser", "broken", serverAddress, remoteFile, destination)
						Expect(err).To(MatchError(ContainSubstring("handshake failed")))
					})
				})

				Context("when the remote file does not exist", func() {
					BeforeEach(func() {
						err := os.Remove(remoteFile)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						err := sshClient.Download("testuser", "testpass", serverAddress, remoteFile, destination)
						Expect(err).To(MatchError(ContainSubstring("file does not exist")))
					})
				})

				Context("when the writer fails", func() {
					var w io.Writer

					BeforeEach(func() {
						pr, pw := io.Pipe()
						pr.Close()
						w = pw
					})

					It("returns an error", func() {
						err := sshClient.Download("testuser", "testpass", serverAddress, remoteFile, w)
						Expect(err).To(MatchError(ContainSubstring("write on closed pipe")))
					})
				})
			})

			Describe("DownloadFile", func() {
				var localFile string

				BeforeEach(func() {
					localFile = filepath.Join(localDir, "local-file")
				})

				It("creates the local file and writes the contents of the remote file to it", func() {
					err := sshClient.DownloadFile("testuser", "testpass", serverAddress, remoteFile, localFile)
					Expect(err).NotTo(HaveOccurred())

					localContents, err := ioutil.ReadFile(localFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(localContents).To(Equal([]byte("file contents\non two lines")))
				})

				Context("when the local file already exists", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(localFile, []byte("some garbage"), 0644)
						Expect(err).NotTo(HaveOccurred())
						Expect(localFile).To(BeAnExistingFile())
					})

					It("replaces the contents of the local file with the contents of the remote file to it", func() {
						err := sshClient.DownloadFile("testuser", "testpass", serverAddress, remoteFile, localFile)
						Expect(err).NotTo(HaveOccurred())

						localContents, err := ioutil.ReadFile(localFile)
						Expect(err).NotTo(HaveOccurred())
						Expect(localContents).To(Equal([]byte("file contents\non two lines")))
					})
				})
			})
		})

		Describe("uploads", func() {
			var remoteFile string

			BeforeEach(func() {
				remoteFile = filepath.Join(remoteDir, "upload-target")
			})

			Describe("Upload", func() {
				var source io.Reader

				BeforeEach(func() {
					source = strings.NewReader("file contents\non two lines")
				})

				It("writes the contents of the source to the remote file", func() {
					err := sshClient.Upload("testuser", "testpass", serverAddress, source, remoteFile)
					Expect(err).NotTo(HaveOccurred())

					targetContents, err := ioutil.ReadFile(remoteFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(targetContents).To(Equal([]byte("file contents\non two lines")))
				})

				Context("when authentication fails", func() {
					It("returns an error", func() {
						err := sshClient.Upload("testuser", "broken", serverAddress, source, remoteFile)
						Expect(err).To(MatchError(ContainSubstring("handshake failed")))
					})
				})

				Context("when target file is a directory", func() {
					It("returns an error", func() {
						err := sshClient.Upload("testuser", "testpass", serverAddress, source, remoteDir)
						Expect(err).To(MatchError(ContainSubstring("is a directory")))
					})
				})

				Context("when target file exists", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(remoteFile, []byte("some garbage"), 0644)
						Expect(err).NotTo(HaveOccurred())
						Expect(remoteFile).To(BeAnExistingFile())
					})

					It("replaces the contents of the file with the contents of the source", func() {
						err := sshClient.Upload("testuser", "testpass", serverAddress, source, remoteFile)
						Expect(err).NotTo(HaveOccurred())

						targetContents, err := ioutil.ReadFile(remoteFile)
						Expect(err).NotTo(HaveOccurred())
						Expect(targetContents).To(Equal([]byte("file contents\non two lines")))
					})
				})

				Context("when the source fails", func() {
					BeforeEach(func() {
						pr, _ := io.Pipe()
						pr.Close()
						source = pr
					})

					It("returns an error", func() {
						err := sshClient.Upload("testuser", "testpass", serverAddress, source, remoteFile)
						Expect(err).To(MatchError(ContainSubstring("closed pipe")))
					})
				})
			})

			Describe("UploadFile", func() {
				var localFile string

				BeforeEach(func() {
					localFile = filepath.Join(localDir, "upload-source")
					err := ioutil.WriteFile(localFile, []byte("file contents\non two lines"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("writes the contents of the local file to the remote file", func() {
					err := sshClient.UploadFile("testuser", "testpass", serverAddress, localFile, remoteFile)
					Expect(err).NotTo(HaveOccurred())

					targetContents, err := ioutil.ReadFile(remoteFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(targetContents).To(Equal([]byte("file contents\non two lines")))
				})

				Context("when the local file does not exist", func() {
					BeforeEach(func() {
						localFile = filepath.Join(localDir, "non-existent")
					})

					It("returns an error", func() {
						err := sshClient.UploadFile("testuser", "testpass", serverAddress, localFile, remoteDir)
						Expect(err).To(HaveOccurred())
						Expect(os.IsNotExist(err)).To(BeTrue())
					})
				})
			})
		})
	})
})

func newListener() (net.Listener, string) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())

	return listener, listener.Addr().String()
}

type server struct {
	mutex        *sync.Mutex
	listener     net.Listener
	serverConfig *ssh.ServerConfig

	sessionChannelHandler func(ssh.NewChannel)

	stopping bool
}

func (s *server) Serve() {
	l := s.listener
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			if s.isStopping() {
				break
			}
			continue
		}

		go s.handleConnection(conn)

		if s.isStopping() {
			break
		}
	}
}

func (s *server) handleConnection(conn net.Conn) {
	serverConn, serverChannels, serverRequests, err := ssh.NewServerConn(conn, s.serverConfig)
	if err != nil {
		return
	}
	defer serverConn.Close()

	go ssh.DiscardRequests(serverRequests)
	go s.handleNewChannels(serverChannels)

	serverConn.Wait()
}

func (s *server) handleNewChannels(newChannelRequests <-chan ssh.NewChannel) {
	for newChannel := range newChannelRequests {
		switch newChannel.ChannelType() {
		case "session":
			go s.sessionChannelHandler(newChannel)
		default:
			newChannel.Reject(ssh.UnknownChannelType, newChannel.ChannelType())
		}
	}
}

func (s *server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stopping = true
	s.listener.Close()
}

func (s *server) isStopping() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.stopping
}

const hostKey = `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`
