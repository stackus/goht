package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/stackus/errors"
	"go.lsp.dev/jsonrpc2"

	"github.com/stackus/goht/internal/logging"
	"github.com/stackus/goht/internal/protocol"
	"github.com/stackus/goht/internal/proxy"
)

type lspFlags struct {
	logFile     string
	traceClient bool
	traceGoPls  bool
}

var lspOptions lspFlags

type rwc struct {
	r io.ReadCloser
	w io.WriteCloser
}

// lspCmd represents the protocol command
var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Starts a language server for goht files",
	Long: `Starts a language server for goht files. The language server provides
code completion, hover tooltips, and more. See https://langserver.org/ for
more information about language servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLsp()
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)

	lspCmd.Flags().StringVar(&lspOptions.logFile, "logFile", "", "log to a file (default stderr)")
	lspCmd.Flags().BoolVar(&lspOptions.traceClient, "traceClient", false, "trace the language server communication")
	lspCmd.Flags().BoolVar(&lspOptions.traceGoPls, "traceGoPls", false, "trace the gopls communication")
}

func runLsp() error {
	var logger zerolog.Logger

	if lspOptions.logFile != "" {
		logFile, err := os.OpenFile(lspOptions.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to open file for logging: %w", err)
		}
		logger = logging.NewLogger(logFile, zerolog.InfoLevel)
	} else {
		logger = logging.NewLogger(os.Stderr, zerolog.ErrorLevel)
	}

	logger.Info().Msg("starting goht-lsp")

	conn := jsonrpc2.NewConn(func() jsonrpc2.Stream {
		stream := jsonrpc2.NewStream(rwc{
			r: os.Stdin,
			w: os.Stdout,
		})
		if lspOptions.traceClient {
			stream = logging.LoggedStream{
				Label:  "GOHT-LSP",
				Stream: stream,
				Logger: logger,
			}
		}
		return stream
	}())
	defer func(conn jsonrpc2.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("error closing connection with client: %s", err)
		}
	}(conn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	goPlsRWC, err := findAndStartGoPls(ctx)
	if err != nil {
		return err
	}
	goConn := jsonrpc2.NewConn(func() jsonrpc2.Stream {
		stream := jsonrpc2.NewStream(goPlsRWC)
		if lspOptions.traceClient {
			stream = logging.LoggedStream{
				Label:  "GO-LSP",
				Stream: stream,
				Logger: logger,
			}
		}
		return stream
	}())
	defer func(goConn jsonrpc2.Conn) {
		err := goConn.Close()
		if err != nil {
			log.Errorf("error closing connection with gopls: %s", err)
		}
	}(goConn)

	client := protocol.ClientDispatcher(conn)
	server := protocol.ServerDispatcher(goConn)

	smc := proxy.NewSourceMapCache()
	dc := proxy.NewDiagnosticsCache()
	srcs := proxy.NewDocumentContents()

	proxyClient := proxy.NewClient(client, smc, dc, logger)
	goConn.Go(
		ctx,
		protocol.Handlers(
			protocol.ClientHandler(proxyClient, jsonrpc2.MethodNotFoundHandler),
		),
	)

	proxyServer := proxy.NewServer(server, client, smc, dc, srcs, logger)
	conn.Go(
		ctx,
		protocol.Handlers(
			protocol.ServerHandler(proxyServer, jsonrpc2.MethodNotFoundHandler),
		),
	)

	select {
	case <-ctx.Done():
		logger.Error().Err(err).Msg("error: context canceled")
		return ctx.Err()
	case <-conn.Done():
		if err := conn.Err(); err != nil {
			logger.Error().Err(err).Msg("error: conn with the client closed")
		}
		return conn.Err()
	case <-goConn.Done():
		if err := goConn.Err(); err != nil {
			logger.Error().Err(err).Msg("error: conn with gopls closed")
		}
		return goConn.Err()
	}
}

// findAndStartGoPls locates the gopls executable and returns a ReadWriteCloser to it
func findAndStartGoPls(ctx context.Context) (io.ReadWriteCloser, error) {
	if _, err := exec.LookPath("gopls"); err != nil {
		return nil, errors.ErrNotFound.Wrap(err, "cannot find gopls in the path. Install gopls with `go install golang.org/x/tools/gopls@latest` and make sure it's in your PATH.")
	}

	cmd := exec.CommandContext(ctx, "gopls")

	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	rc := &rwc{
		r: out,
		w: in,
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("gopls exited with error: %s", err)
		}
	}()

	return rc, nil
}

func (rwc rwc) Read(b []byte) (int, error)  { return rwc.r.Read(b) }
func (rwc rwc) Write(b []byte) (int, error) { return rwc.w.Write(b) }
func (rwc rwc) Close() error {
	rwc.r.Close()
	return rwc.w.Close()
}
