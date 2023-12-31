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
document.getElementById("download").addEventListener("click", function(){
    // Generate download of csv file with content
    var filename = document.getElementById("filename").value;
    var text = document.getElementById("filecontent").value;
    download(filename, text);
}, false);

/* 
    download plain text UTF8 file 
    NEEDS:
        - 1 button with id="download"
        - 1 text element with id="filename"
        - 1 text element with id="filecontent"
*/