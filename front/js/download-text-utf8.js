function download(filename, text) {
    var element = document.createElement('a');
    element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
    element.setAttribute('download', filename);

    element.style.display = 'none';
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
}

// Start file download.
// check JS before submitting the HTMX request
document.body.addEventListener("htmx:confirm", function(evt){
    evt.preventDefault();
    if (evt.detail.path != "/export-csv-download"){
        evt.detail.issueRequest();
    } else {
        fileContent = document.getElementById("filecontent").value
        if (fileContent.substring(0, 4) == "Rien"){
            window.alert("Rien à télécharger");
        } else {
            // Generate download of csv file with content
            var filename = document.getElementById("filename").value;
            // "\uFEFF" = BOM : to generate an UTF-8 file with BOM
            var text = "\uFEFF" + fileContent;
            download(filename, text);
            evt.detail.issueRequest();
        }
    }
}, false);

/* 
    download plain text UTF8 file 
    NEEDS:
        - 1 button with id="download"
        - 1 text element with id="filename"
        - 1 text element with id="filecontent"
*/