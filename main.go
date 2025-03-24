package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	p := &proxy{}
	http.HandleFunc("/", p.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local testing
	}

	fmt.Printf("Starting proxy on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		panic(err)
	}
}

type proxy struct{}

func (p *proxy) Handler(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://www.aezenai.com" + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Failed to fetch content", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)

	if !strings.Contains(contentType, "text/html") {
		w.WriteHeader(resp.StatusCode)
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, "Failed to serve content", http.StatusInternalServerError)
		}
		return
	}

	modifiedHTML := p.injectJS(resp.Body)
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(modifiedHTML)
	if err != nil {
		http.Error(w, "Failed to write modified HTML", http.StatusInternalServerError)
	}
}

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

    // Enhanced removal of <div class="AuthLayout_left__mnQq_ flex-1">
    let authLayoutDivs = document.querySelectorAll('div[class*="AuthLayout_left__"][class*="flex-1"]');
    authLayoutDivs.forEach(div => {
        if (div.className.includes("AuthLayout_left__") && div.className.includes("flex-1")) {
            div.remove();
        }
    });

    // Remove <div class="Text_text__0_Dq5 Text_title_2__yppuO"> when it contains "Referral"
    let referralTextDivs = document.querySelectorAll('div.Text_text__0_Dq5.Text_title_2__yppuO');
    referralTextDivs.forEach(div => {
        if (div.textContent.trim() === "Referral") {
            console.log("Found Text_text__0_Dq5 Text_title_2__yppuO div with 'Referral', removing:", div);
            div.remove();
        }
    });

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
        "/workspaces/7368da06-9cc6-4e7c-82f8-207da38b5e12/bots/af2a896a-5517-48eb-b8f3-014eec338f38?t=platform-integration",
        "/user-settings?workspace_id=7368da06-9cc6-4e7c-82f8-207da38b5e12&t=referral",
        "https://drive.google.com/file/d/1ZOdQl4HIhsOGyV7X5urKUh8kNdUuh8k4/view"
    ];
    hrefsToRemove.forEach(href => {
        let aToRemove = document.querySelector('a[href="' + href + '"]');
        if (aToRemove) {
            console.log("Found <a> with href " + href + ", removing:", aToRemove);
            aToRemove.remove();
        }
    });

    // Remove <span> with class "icon-message-question"
    let messageQuestionSpans = document.querySelectorAll('span.icon-message-question');
    messageQuestionSpans.forEach(span => {
        console.log("Found span with class icon-message-question, removing:", span);
        span.remove();
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

    // Remove <div class="ChatButton_get_started_button__h0pJh react-draggable">
    let chatButtonDiv = document.querySelector('div.ChatButton_get_started_button__h0pJh.react-draggable');
    if (chatButtonDiv) chatButtonDiv.remove();

    // Remove <h2> with "Application error: a client-side exception has occurred"
    let errorH2 = Array.from(document.querySelectorAll('h2')).find(h2 => h2.textContent.includes("Application error: a client-side exception has occurred"));
    if (errorH2) errorH2.remove();

    // Remove <div class="Text_text__0_Dq5 Text_title_2__yppuO text-center max-w-[200px]" with text "cxgenie_works_best_on_desktop_version">
    let cxGenieDivs = document.querySelectorAll('div.Text_text__0_Dq5.Text_title_2__yppuO.text-center.max-w-\\[200px\\]');
    Array.from(cxGenieDivs).forEach(div => {
        if (div.textContent.trim().includes("cxgenie_works_best_on_desktop_version")) {
            console.log("Found cxgenie_works_best_on_desktop_version div, removing:", div);
            div.remove();
        }
    });

    // Remove <div class="Sidebar_trained_data__LCzjK">
    let sidebarTrainedDataDivs = document.querySelectorAll('div.Sidebar_trained_data__LCzjK');
    sidebarTrainedDataDivs.forEach(div => {
        console.log("Found Sidebar_trained_data__LCzjK div, removing:", div);
        div.remove();
    });

    // Remove <div class="h-[38px] flex items-center justify-between px-3" data-sentry-component="ReferralCode">
    let referralCodeDivs = document.querySelectorAll('div.h-\\[38px\\].flex.items-center.justify-between.px-3[data-sentry-component="ReferralCode"]');
    referralCodeDivs.forEach(div => {
        console.log("Found ReferralCode div, removing:", div);
        div.remove();
    });

    // Remove all <div data-node-key="referral" class="cxg-tabs-tab">
    let referralTabDivs = document.querySelectorAll('div[data-node-key="referral"].cxg-tabs-tab');
    referralTabDivs.forEach(div => {
        console.log("Found referral tab div, removing:", div);
        div.remove();
    });

    // Remove <div class="flex items-center gap-2"> containing "Referral" and a span with class "icon-share"
    let referralIconDivs = Array.from(document.querySelectorAll('div.flex.items-center.gap-2')).filter(div => {
        return div.textContent.includes("Referral") && div.querySelector('span.icon-share');
    });
    referralIconDivs.forEach(div => {
        console.log("Found div with class flex items-center gap-2 containing Referral and icon-share, removing:", div);
        div.remove();
    });
}

// Function to check for error in title and add reload button
function checkForErrorAndAddButton() {
    const title = document.title;
    if (title.includes("Application error: a client-side exception has occurred")) {
        // Make whole screen white
        document.body.style.backgroundColor = 'white';
        document.body.innerHTML = ''; // Clear existing content

        if (!document.getElementById('reload-button')) {
            const button = document.createElement('button');
            button.id = 'reload-button';
            button.textContent = 'Return Back';
            button.style.position = 'fixed';
            button.style.top = '50%';
            button.style.left = '50%';
            button.style.transform = 'translate(-50%, -50%)';
            button.style.zIndex = '1000';
            button.style.padding = '10px 20px';
            button.style.fontSize = '16px';
            button.style.cursor = 'pointer';
            button.style.backgroundColor = '#007bff'; // Blue background
            button.style.color = 'white'; // White text
            button.style.border = 'none';
            button.style.borderRadius = '4px';
            button.addEventListener('click', () => {
                window.location.href = 'https://portal-aezenai.onrender.com/';
            });
            document.body.appendChild(button);
        }
    }
}

// Run initially
removeUnwantedElements();
checkForErrorAndAddButton();

// MutationObserver to detect new elements in body
const bodyObserver = new MutationObserver(() => removeUnwantedElements());
bodyObserver.observe(document.body, { childList: true, subtree: true });

// MutationObserver to detect changes in head (for title updates)
const headObserver = new MutationObserver(() => checkForErrorAndAddButton());
headObserver.observe(document.head, { childList: true, subtree: true });

// Fallback interval for extra reliability on element removal
setInterval(removeUnwantedElements, 100);
				`,
			},
		}
		n.AppendChild(script)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		injectScript(c)
	}
}
