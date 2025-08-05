package main_test

import (
	"bufio"
	"log/slog"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/onsi/biloba"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVncWsGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vnc-ws-gateway E2E Suite")
}

var _ = BeforeSuite(func() {
	slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{Level: slog.LevelDebug})))

	gateway := exec.Command("bin/vnc-ws-gateway", "--vnc-poll-period", "1s")
	gateway.Env = append(gateway.Env, os.Environ()...)
	gateway.Env = append(gateway.Env, "VNC_WS_GATEWAY_PASSWORD=password")
	stdout, err := gateway.StdoutPipe()
	Expect(err).ToNot(HaveOccurred())
	go func() {
		scan := bufio.NewScanner(stdout)
		for scan.Scan() {
			GinkgoWriter.Println(scan.Text())
		}
	}()
	stderr, err := gateway.StderrPipe()
	Expect(err).ToNot(HaveOccurred())
	go func() {
		scan := bufio.NewScanner(stderr)
		for scan.Scan() {
			GinkgoWriter.Println(scan.Text())
		}
	}()
	slog.Info("starting gateway process")
	Expect(gateway.Start()).To(Succeed())
	DeferCleanup(func() {
		slog.Info("signaling gateway to stop")
		Expect(gateway.Process.Signal(os.Interrupt)).To(Succeed())
		slog.Info("waiting for gateway to stop")
		gateway.Wait()
	})

	bopts := []chromedp.ExecAllocatorOption{
		// chromedp.ProxyServer(proxyURL),
		// chromedp.Flag("headless", false),
		// chromedp.Flag("ignore-certificate-errors", "1"),
		chromedp.WindowSize(1920, 1080),
	}
	if runningInContainer {
		bopts = append(bopts, chromedp.NoSandbox)
		GinkgoWriter.Printf("!!! WARNING: Sandbox disabled due to containerized environment detected from %s. This is insecure if this not actually a container!\n", RunningInContainerEnv)
	}

	slog.Info("starting browser")
	biloba.SpinUpChrome(GinkgoT(), bopts...)
})

const (
	RunningInContainerEnv = "RUNNING_IN_CONTAINER"
	DevModeEnv            = "E2E_DEV_MODE"
	IsCIEnv               = "IS_CI"
)

var (
	isCI               = os.Getenv(IsCIEnv) != ""
	runningInContainer = os.Getenv(RunningInContainerEnv) != ""
	devMode            = os.Getenv(DevModeEnv) != ""
)

var _ = Describe("gateway", func() {
	var b *biloba.Biloba
	BeforeEach(func() {
		b = biloba.ConnectToChrome(GinkgoT()).NewTab()
	})
	It("should launch a VNC session in the browser", func() {
		b.HandlePromptDialogs().MatchingMessage("Enter VNC Password").WithText("password")
		b.HandlePromptDialogs().MatchingMessage("Choose display resolution\n1920x1200, 1920x1080, 1600x1200, 1680x1050, 1400x1050, 1360x768, 1280x1024, 1280x960, 1280x800, 1280x720, 1024x768, 800x600, 640x480").WithText("1920x1080")
		b.HandleAlertDialogs().WithResponse(true)
		slog.Info("navigating to page")
		b.Navigate("http://localhost:8080")

		Eventually(`body canvas`, "30s").Should(b.Exist())
		// TODO: Find out what actually triggers this to exist
		// b.Click(`body canvas`)
		// Eventually(`body div#noVNC_mouse_capture_elem`, "5s").Should(b.Exist())
		Eventually(`body canvas`, "5s").Should(b.HaveProperty("width", 1920.0))
		Expect(`body canvas`).To(b.HaveProperty("height", 1080.0))
		// TODO: Find/make dummy graphical app to test sending keystrokes using raw chromedp

		b.Close()
	})

	It("should launch a VNC session in the browser when proxied at a subpath", func() {

		rootPath := "/a/root/subpath/"
		proxy := httptest.NewServer(&httputil.ReverseProxy{
			Rewrite: func(req *httputil.ProxyRequest) {
				defer GinkgoRecover()
				url := *req.In.URL
				url.Scheme = "http"
				url.Host = "localhost:8080"
				slog.Info("???", "path", url.Path, "prefix", rootPath, "trimmed", strings.TrimPrefix(url.Path, rootPath))
				url.Path = strings.TrimPrefix(url.Path, rootPath)
				if !strings.HasPrefix(url.Path, "/") {
					url.Path = "/" + url.Path
				}
				req.Out.URL = &url
				slog.Info("proxying", "in", req.In.URL, "out", req.Out.URL, "url", &url)
			},
		})
		DeferCleanup(proxy.Close)

		b.HandlePromptDialogs().MatchingMessage("Enter VNC Password").WithText("password")
		b.HandlePromptDialogs().MatchingMessage("Choose display resolution\n1920x1200, 1920x1080, 1600x1200, 1680x1050, 1400x1050, 1360x768, 1280x1024, 1280x960, 1280x800, 1280x720, 1024x768, 800x600, 640x480").WithText("1920x1080")
		b.HandleAlertDialogs().WithResponse(true)
		slog.Info("navigating to page")
		b.Navigate(proxy.URL + rootPath)

		Eventually(`body canvas`, "30s").Should(b.Exist())
		// TODO: Find out what actually triggers this to exist
		// b.Click(`body canvas`)
		// Eventually(`body div#noVNC_mouse_capture_elem`, "5s").Should(b.Exist())
		Eventually(`body canvas`, "5s").Should(b.HaveProperty("width", 1920.0))
		Expect(`body canvas`).To(b.HaveProperty("height", 1080.0))
		// TODO: Find/make dummy graphical app to test sending keystrokes using raw chromedp

		b.Close()
	})
})
