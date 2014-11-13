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
        xhr.setRequestHeader("Accept", "application/json");
        xhr.open("GET", document.URL, true);
        xhr.onreadystatechange = function () {
        	if (xhr.status !== 4) {
        		return;
        	}
        	console.log(xhr.response);
        };
    };

    return {
        init: function () {
        	setTimeout(check, 1000);
        }
    };
}());