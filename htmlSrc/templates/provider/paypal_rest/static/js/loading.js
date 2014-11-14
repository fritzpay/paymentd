/*
 * Get cross browser xhr object
 *
 *
 *            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
 *                    Version 2, December 2004
 *
 * Copyright (C) 2011 Jed Schmidt <http://jed.is>
 * More: https://gist.github.com/993585
 *
 */
var j = function (a) {
    "use strict";
    var ms = ["Msxml2", "Msxml3", "Microsoft"];
    for (a = 0; a < 4; a += 1) {
        try {
            return a ? new ActiveXObject(ms[a] + ".XMLHTTP") : new XMLHttpRequest();
        } catch (ignore) {}
    }
};

var loading = (function () {
    "use strict";
    var check = function () {
        var xhr = j();
        xhr.open("GET", document.URL, true);
        xhr.setRequestHeader("Accept", "application/json");
        xhr.responseType = "json";
        xhr.onreadystatechange = function () {
        	if (xhr.readyState !== 4) {
        		return;
        	}
            if (xhr.status !== 200) {
                location.reload();
                return;
            }
            if (xhr.response.c !== undefined && xhr.response.c === true) {
                location.reload();
                return;
            }
            setTimeout(check, 1000);
        };
        xhr.send();
    };

    return {
        init: function () {
        	setTimeout(check, 1000);
        }
    };
}());
loading.init();
