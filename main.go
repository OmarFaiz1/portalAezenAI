package main

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	p := &proxy{}
	http.HandleFunc("/", p.Handler)

	// Get the port from the PORT environment variable (required by Render)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback to 8080 for local testing
	}

	// Listen on the assigned port
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

type proxy struct{}

func (p *proxy) Handler(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://www.aezenai.com" + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery // Preserve query parameters if present
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Failed to fetch content", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy all headers from the target response to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set Content-Type explicitly for non-HTML content (e.g., images)
	contentType := resp.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)

	// If it's not an HTML page (e.g., images, JS, CSS), serve it directly
	if !strings.Contains(contentType, "text/html") {
		w.WriteHeader(resp.StatusCode)
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, "Failed to serve content", http.StatusInternalServerError)
		}
		return
	}

	// For HTML pages, modify and inject JavaScript
	modifiedHTML := p.injectJS(resp.Body)
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(modifiedHTML)
	if err != nil {
		http.Error(w, "Failed to write modified HTML", http.StatusInternalServerError)
	}
}

// injectJS injects JavaScript to remove unwanted elements and disable right-click
func (p *proxy) injectJS(body io.Reader) []byte {
	doc, err := html.Parse(body)
	if err != nil {
		return []byte("Error parsing HTML")
	}

	var buf bytes.Buffer
	injectScript(doc)
	err = html.Render(&buf, doc)
	if err != nil {
		return []byte("Error rendering HTML")
	}

	return buf.Bytes()
}

// injectScript adds a JavaScript snippet to remove elements and disable right-click
func injectScript(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "body" {
		script := &html.Node{
			Type: html.ElementNode,
			Data: "script",
			FirstChild: &html.Node{
				Type: html.TextNode,
				Data: `
function removeUnwantedElements() {
    // Remove <div> with specific texts
    let divTextsToRemove = [
        "CX Genie works best on desktop version",
        "Basic information suggested by CX Genie"
    ];
    let divsToRemove = Array.from(document.querySelectorAll('div')).filter(div => divTextsToRemove.includes(div.textContent.trim()));
    divsToRemove.forEach(div => div.remove());

    // Remove <div class="AuthLayout_left__mnQq_ flex-1"> containing infinite scroll
    let authLayoutDiv = document.querySelector('div.AuthLayout_left__mnQq_.flex-1');
    if (authLayoutDiv) authLayoutDiv.remove();

    // Remove <div> with watermark toggle
    let watermarkDivToRemove = Array.from(document.querySelectorAll('div')).find(div => {
        let titles = Array.from(div.querySelectorAll('div')).filter(d => d.textContent.trim() === "Watermark");
        let descriptions = Array.from(div.querySelectorAll('div')).filter(d => d.textContent.trim() === "Show / hide watermark on chat widget");
        return titles.length > 0 && descriptions.length > 0;
    });
    if (watermarkDivToRemove) watermarkDivToRemove.remove();

    // Remove <span> with class "hidden xl:inline" and specific texts
    let spanTextsToRemove = ["How should I add questions?", "How should I add the articles?"];
    let spansToRemove = Array.from(document.querySelectorAll('span.hidden.xl\\:inline')).filter(span => spanTextsToRemove.includes(span.textContent.trim()));
    spansToRemove.forEach(span => span.remove());

    // Remove standalone <span> with class "icon-info-circle-2"
    let iconSpansToRemove = document.querySelectorAll('span.icon-info-circle-2');
    iconSpansToRemove.forEach(span => span.remove());

    // Remove <a> with specific hrefs
    let hrefsToRemove = [
        "https://drive.google.com/file/d/1ulFvSQUYmfXNCulWC76_HEbync67tAYH/view?usp=sharing",
        "https://drive.google.com/file/d/1Lnnm5vppjx27QayOn-7eMcGlnal52tv3/view?usp=sharing",
        "/workspaces/7368da06-9cc6-4e7c-82f8-207da38b5e12/bots/af2a896a-5517-48eb-b8f3-014eec338f38?t=platform-integration"
    ];
    hrefsToRemove.forEach(href => {
        let aToRemove = document.querySelector('a[href="' + href + '"]');
        if (aToRemove) aToRemove.remove();
    });

    // Remove <button> that contains a <span> with specific texts
    let buttonsToRemove = Array.from(document.querySelectorAll('button')).filter(button => {
        let spans = button.querySelectorAll('span');
        return Array.from(spans).some(span => spanTextsToRemove.includes(span.textContent.trim()));
    });
    buttonsToRemove.forEach(button => button.remove());

    // Remove <button> with trash icon
    let trashButtons = Array.from(document.querySelectorAll('button')).filter(button => button.querySelector('span.icon-trash'));
    trashButtons.forEach(button => button.remove());
}

// Run initially
removeUnwantedElements();

// MutationObserver to detect new elements
const observer = new MutationObserver(() => removeUnwantedElements());
observer.observe(document.body, { childList: true, subtree: true });

// Fallback interval for extra reliability
setInterval(removeUnwantedElements, 100);

// Disable right-click (context menu)
document.addEventListener('contextmenu', function(e) {
    e.preventDefault();
});
				`,
			},
		}
		n.AppendChild(script)
	}

	// Recursively apply to all child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		injectScript(c)
	}
}