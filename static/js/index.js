var M = {};

M.warningAlert = function() {}
M.warningAlert.warning = function(message) {
            $("#alert-container").html('<div class="alert alert-danger"><a class="close" data-dismiss="alert">×</a><span>'+message+'</span></div>')
        }
M.infoAlert = function() {}
M.infoAlert.info = function(message) {
            $("#alert-container").html('<div class="alert alert-info"><a class="close" data-dismiss="alert">×</a><span>'+message+'</span></div>')
        }

$(document).ready(function() {
    $("#memorize-modal").on("shown.bs.modal", function() {
        $("#input-email").focus();
    })
});

function handleSearchSubmit(event) {
    event.preventDefault();
        performSearch();
    return false;
}

function handleEmailSubmit(event) {
    event.preventDefault();
    performAdd();
    return false;
}

$(function () {
    $(".dropdown-menu").on("click", "button", function () {
        var btn = $(this).parents(".dropdown").find(".btn");
        var orgVal = btn.data("value");
        var newVal = $(this).data("value");
        if(orgVal != newVal) {
            btn.data("value", newVal);
            performSearch();
        }
    });
});

// Used to detect initial (useless) popstate.
// If history.state exists, assume browser isn't going to fire initial popstate.
var popped = ('state' in window.history && window.history.state !== null), initialURL = location.href;

$(window).on("popstate", function (event) {
    // Ignore inital popstate that some browsers fire on page load
    var initialPop = !popped && location.href == initialURL
    popped = true
    if (initialPop) {
        return;
    }

    if ("Memorize It" === document.title) {
        window.location.replace(document.location);
    }
});

function disableModalClose() {
    $(".modal.fade").data("keyboard", "false");
    $("#input-email").prop('disabled', true);
    $("#add-email").prop('disabled', true);
    $(".close").prop('disabled', true);
}

function enableModalClose() {
    $(".modal.fade").data("keyboard", "true");
    $("#input-email").prop('disabled', false);
    $("#add-email").prop('disabled', false);
    $(".close").prop('disabled', false);
    $("#memorize-modal").modal("hide");
}

function performAdd() {
    var headword = $(".inner").data("headword");
    if (headword=="") {
        return;
    }
    disableModalClose();
    $("#add-email").text("Memorizing...");
    var email = $("#input-email").val();
    var type = $(".dropdown").find(".btn").data("value");
    $.ajax({
        type: "POST",
        url: "/api/add/" + email + "/" + type + "/" + headword,
        dataType: "json",
        success: function (r) {
            switch (r.result) {
                case "Success":
                M.infoAlert.info("Reminders will be sent for this word to your email.");
                break;
                case "AlreadyExist":
                M.warningAlert.warning("You are already memorizing this word."); //TODO maybe reset the memorization state?
                break;
                case "WordNotFound":
                break;
                case "InvalidInputs":
                case "UserNotFound":
                case "Error":
                //should never get here
                break;
            }
        },
        error: function (xhr, status, text) {
            var r = JSON.parse(xhr.responseText);
            M.warningAlert.warning("An error is encountered: "+r.error);
        },
        complete: function() {
            enableModalClose();
            $("#add-email").text("Memorize");
        }
    });
}

function performSearch() {
    var type = $(".dropdown").find(".btn").data("value");
    var headword = $("#input-headword").val();
    if (headword == "") {
        headword = $(".inner").data("headword");
    }
    doSearch(type, headword);
}

function doSearch(type, headword) {
    if (headword == "") {
        return;
    }
    window.location.replace("/def/" + type + "/" + headword);
    history.pushState(null, "Memorize It", "/def/" + type + "/" + headword);
//    window.location.href = "/def/" + type + "/" + headword;
    return;
//    $.ajax({
//        type: "GET",
//        url: "/api/def/" + type + "/" + headword,
//        dataType: "json",
//        success: function (r) {
//        console.log(r.valid);
//            if (r.valid) {
//                for (var i = 0; i < r.results.length; i++) {
//                    alert(r.results[i].part_of_speech);
//                    alert(r.results[i].definition);
//                    if (r.results[i].synonyms != undefined) {
//                        alert(r.results[i].synonyms[0]);
//                    }
//                    alert(r.results[i].examples);
//                }
//              $(".inner").data("headword", r.headword);
//            } else {
//                alert("not found");
//            }
//        },
//        error: function (xhr, status, text) {
//            var r = JSON.parse(xhr.responseText);
//            if (r.result == "InvalidInputs") {
//                alert("Invalid input(s)");
//            } else {
//                alert(r.result);
//            }
//        }
//    });
}