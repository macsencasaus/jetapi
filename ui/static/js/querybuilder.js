const regField = document.getElementById("reg");
regField.value = "";

const queryUrl = document.getElementById("query_url");
const photosInput = document.getElementById("photos_input");
const flightsInput = document.getElementById("flights_input");
const onlyJPInput = document.getElementById("only_jp");
const onlyFRInput = document.getElementById("only_fr");

const apiUrl = `${window.location.origin}/api?reg=`;

photosInput.value = 3;
flightsInput.value = 20;
queryUrl.textContent = apiUrl;

document.getElementById("main").addEventListener("input", function (_) {
    const reg = regField.value;
    photos = photosInput.value;
    flights = flightsInput.value;
    query = apiUrl + reg;

    function clamp(n) {
        return Math.min(20, Math.max(0, n));
    }

    photos = clamp(photos);
    photosInput.value = clamp(photosInput.value);

    flights = clamp(flights);
    flightsInput.value = clamp(flightsInput.value);

    if (photos != 3) {
        query += "&photos=" + photos;
    }
    if (flights != 20) {
        query += "&flights=" + flights;
    }
    if (onlyJPInput.checked) {
        query += "&only_jp=true";
    }
    if (onlyFRInput.checked) {
        query += "&only_fr=true";
    }

    if (regField.value != "") {
        queryUrl.textContent = query;
    } else {
        queryUrl.textContent = apiUrl;
    }
});

function copy(copyButton, copyId) {
    copyButton.addEventListener("click", function (_) {
        const selection = window.getSelection();
        const range = document.createRange();
        range.selectNodeContents(copyId);
        selection.removeAllRanges();
        selection.addRange(range);

        document.execCommand("copy");

        selection.removeAllRanges();
    });
}

const copyButton = document.getElementById("copy");
copy(copyButton, queryUrl);

const getButton = document.getElementById("get");
const jsonDiv = document.getElementById("json");

getButton.addEventListener("click", function (_) {
    reg = regField.value;
    jsonDiv.innerHTML = "<p>Loading...</p>";
    fetch(queryUrl.textContent)
        .then((response) => {
            if (!response.ok) {
                jsonDiv.innerHTML = `
                    <p>Error</p>
                `;
                throw new Error("Network reponse was not ok");
            }
            return response.json();
        })
        .then((data) => {
            const jsonString = JSON.stringify(data, null, 2);
            jsonDiv.innerHTML = ` 
                <div class="json_output">
                    <pre id="raw_json">${jsonString}</pre>
                    <div class="json_options">
                        <button type="submit" class="copy" id="copy_json">
                            <img src="/static/img/copy.png">
                        </button>
                        <button type="submit" class="download" id="download_json">
                            <img src="/static/img/download.png">
                        </button>
                    </div>
                </div>
            `;
            const copyJSONButton = document.getElementById("copy_json");
            copy(copyJSONButton, document.getElementById("raw_json"));
            const downloadJSONButton = document.getElementById("download_json");
            downloadJSONButton.onclick = function () {
                const blob = new Blob([jsonString], {
                    type: "application/json",
                });
                const downloadURL = URL.createObjectURL(blob);
                const link = document.createElement("a");
                link.href = downloadURL;
                link.download = `${reg}.json`;
                document.body.appendChild(link);
                link.click();
                document.body.removeChild(link);
                URL.revokeObjectUrl(downloadURL);
            };
        });
});
