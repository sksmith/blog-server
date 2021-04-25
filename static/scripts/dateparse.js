function parseDate(d) {
    return new Date(Date.parse(d));
}

function localizeTimes() {
    var list = document.getElementsByTagName("time");
    Array.prototype.slice.call(list).forEach(function(element) {
        let datetime = element.getAttribute("datetime");
        parsedDate = parseDate(datetime);
        element.innerHTML = formatDate(parsedDate);
        element.setAttribute("datetime", datetime);
    });
}

let months = new Map();
months[0] = "January";
months[1] = "February";
months[2] = "March";
months[3] = "April";
months[4] = "May";
months[5] = "June";
months[6] = "July";
months[7] = "August";
months[8] = "September";
months[9] = "October";
months[10] = "November";
months[11] = "December";

function formatDate(d) {
    return months[d.getMonth()] + " " + d.getDate() + ", " + d.getFullYear();
}