function toAppPath(relative) {
    return `/greeter/${relative}` 
}

function isLoggedIn() {
    return !(typeof window.login === "undefined")
}

// Authenticate requests if login token is available
$.ajaxPrefilter(function(options, originalOptions, jqXHR) {
    if (isLoggedIn()) {
        options.headers = {
            "Authorization": `Bearer ${btoa(window.login.access_token)}`
        };
    }
})

// Login
$(document).ready(function() {
    $("#loginPopup").click(function(event) {
        event.preventDefault();
        $("#loginMessage").text(isLoggedIn() ? "Logged in" : "Not logged in");
        $("#loginModal").css("display", "block");
    });

    $("#loginBox").submit(function(event) {
        event.preventDefault();

        var user = $("#loginUser").val();
        var password = $("#loginPassword").val();

        $.ajax({
            type: "POST",
            url: toAppPath("auth/logins"),
            contentType: "application/json",
            data: JSON.stringify({
                user_name: user,
                user_password: password 
            })
        }).fail(function(resp) {
            $("#loginMessage").text(`Login failure: ${resp.status}`);
        }).done(function(resp) {
            // Store jwt in the page root window object
            window.login = {
                user_id: resp.user_id,
                login_id: resp.login_id,
                access_token: resp.access_token,
                refresh_token: resp.refresh_token
            };
            $("#loginMessage").text("Login success");
        });
    });

    $("#loginClose").click(function(event) {
        event.preventDefault();
        $("#loginModal").css("display", "none");
    });
});

// Logout
$(document).ready(function() {
    $("#logoutPopup").click(function(event) {
        event.preventDefault();
        $("#logoutMessage").text(isLoggedIn() ? "Logged in" : "Not logged in");
        $("#logoutModal").css("display", "block");
    });

    $("#logoutBox").submit(function(event) {
        $.ajax({
            type: "DELETE",
            url: toAppPath(`auth/logins/${window.login.login_id}`),
         }).fail(function(resp) {
            $("#logoutMessage").text(`Logout failure: ${resp.status}`);
        }).done(function(resp) {
            $("#logoutMessage").text("Logout success");
        }).always(function() {
            delete window.login
        });
    });

    $("#logoutClose").click(function(event) {
        event.preventDefault();
        $("#logoutModal").css("display", "none");
    });
});

// Create Account
$(document).ready(function() {
    $("#registerPopup").click(function(event) {
        event.preventDefault();
        $("#registerMessage").text(isLoggedIn() ? "Logged in" : "Not logged in");
        $("#registerModal").css("display", "block");
    });

    $("#registerBox").submit(function(event) {
        event.preventDefault();

        var user = $("#registerUser").val();
        var password = $("#registerPassword").val();
        var passwordRepeat = $("#registerPasswordRepeat").val();

        if (!(password === passwordRepeat)) {
            $("#registerMessage").text("Passwords do not match");
            return;
        }

        $.ajax({
            type: "POST",
            url: toAppPath("auth/users"),
            contentType: "application/json",
            data: JSON.stringify({
                user_name: user,
                user_password: password
            })
        }).fail(function(resp) {
            $("#registerMessage").text(`Registration failure: ${resp.status}`);
        }).done(function(resp) {
            $("#registerMessage").text("Registration success");
        });
    });

    $("#registerClose").click(function(event) {
        event.preventDefault();
        $("#registerModal").css("display", "none");
    });
});

// Delete Account
$(document).ready(function() {    
    $("#unregisterPopup").click(function(event) {
        event.preventDefault();
        $("#unregisterMesssage").text(isLoggedIn() ? "Logged in" : "Not logged in");
        $("#unregisterModal").css("display", "block");
    });

    $("#unregisterBox").submit(function(event) {
        event.preventDefault();
        
        $.ajax({
            type: "DELETE",
            url: toAppPath(`auth/users/${window.login.user_id}`),
        }).fail(function(resp) {
            $("#unregisterMessage").text(`Delete account failure: ${resp.status}`);
        }).done(function(resp) {
            $("#unregisterMessage").text("Deleted account");
        }).always(function() {
            delete window.login
        });

        // TODO chain a logout call to this request?
    });

    $("#unregisterClose").click(function(event) {
        event.preventDefault();
        $("#unregisterModal").css("display", "none");
    });
});

// Greeting Preferences
$(document).ready(function() {
    $("#prefsPopup").click(function(event) {
        event.preventDefault();

        // Clear the current prefs content
        $("#prefsLanguages").find("option").remove();
        $("#prefsMessage").text("");

        $.ajax({
            type: "GET",
            url: toAppPath("messages/greetings"),
        }).fail(function(resp) {
            $("#prefsLanguages").append(
                $("<option></option>")
                .attr("value", "none")
                .text(`Failed to get languages: ${resp.status}`));
        }).done(function(resp) {
            $.each(resp.languages, function(key, val) {
                $("#prefsLanguages").append(
                    $("<option></option>")
                    .attr("value", key)
                    .text(key));
            });
        }).always(function() {
            $("#prefsModal").css("display", "block");
        });
    });

    $("#prefsBox").submit(function(event) {
        event.preventDefault();
            
        var language = $("#prefsLanguages").find(":selected").text();
        
        $.ajax({
            type: "PUT",
            url: toAppPath(`messages/users/${window.login.user_id}`),
            contentType: "application/json",
            data: JSON.stringify({
                user_language: language
            })
        }).fail(function(resp) {
            $("#prefsMessage").text(`Failed to save preferences: ${resp.status}`);
        }).done(function(resp) {
            $("#prefsMessage").text("Saved preferences");
        });
    });

    $("#prefsClose").click(function (event) {
        event.preventDefault();
        $("#prefsModal").css("display", "none");
    });
});

// Greeting
$(document).ready(function() {
    $("#greetPopup").click(function(event) {
        event.preventDefault();

        // Clear the current greeter UI content
        $("#greetMessage").text("No greeting");
        
        delete window.prefs;

        $.ajax({
            type: "GET",
            url: toAppPath("messages/users/" + window.login.user_id),
         }).fail(function(resp) {
            $("#greetMessage").text(`Failed to retrieve preferences: ${resp.status}`);
        }).done(function(resp) {
            window.prefs = {
                user_name: resp.user_name,
                user_language: resp.user_language
            };
            $("#greetMessage").text(`Preferences: ${JSON.stringify(window.prefs)}`);
        }).always(function() {
            $("#greetModal").css("display", "block");
        });
    });

    $("#greetBox").submit(function(event) {
        event.preventDefault();

        var date = new Date().toLocaleTimeString();

        $.ajax({
            type: "POST",
            url: toAppPath("messages/greetings"),
            contentType: "application/json",
            data: JSON.stringify({
                "user_id": window.login.user_id,
                "language": window.prefs.user_language
            })
         }).fail(function(resp) {
            $("#greetMessage").text(`(${date}) Failure: ${resp.status}`);
        }).done(function(resp) {
            $("#greetMessage").text(`(${date}) Success: ${resp.message}`);
        });
    });

    $("#greetClose").click(function(event) {
        event.preventDefault();
        $("#greetModal").css("display", "none");
    });
});
