function toAppPath(relative) {
    return "/greeter/" + relative 
}

// Login
$(document).ready(function() {
    $("#loginPopup").click(function(event) {
        event.preventDefault();

        if (typeof window.jwt === "undefined") {
            $("#loginMessage").text("Not logged in");
        } else {
            $("#loginMessage").text("Logged in");
        }

        $("#loginModal").css("display", "block");
    });

    $("#loginBox").submit(function(event) {
        event.preventDefault();

        var user = $("#loginUser").val();
        var password = $("#loginPassword").val();

        $.ajax({
            type: "POST",
            url: toAppPath("auth/login"),
            dataType: "json",
            contentType: "application/json",
            data: JSON.stringify({
                user_name: user,
                user_password: password 
            })
        }).fail(function(resp) {
            $("#loginMessage").text("Login failure: " + resp.status);
        }).done(function(resp) {
            // Store jwt in the page root window object
            window.jwt = resp.access_token;
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

        if (typeof window.jwt === "undefined") {
            $("#logoutMessage").text("Not logged in");
        } else {
            $.ajax({
                type: "POST",
                url: toAppPath("auth/logout"),
                dataType: "json",
                contentType: "application/json",
                data: JSON.stringify({
                    access_token: window.jwt
                })
            }).fail(function(resp) {
                $("#loginMessage").text("Logout failure: " + resp.status);
                delete window.jwt
            }).done(function(resp) {
                $("#loginMessage").text("Logout success");
                delete window.jwt
            });
            
            //$("#logoutMessage").text("Logged out");
        }

        $("#logoutModal").css("display", "block");
    });

    $("#logoutClose").click(function(event) {
        event.preventDefault();
        $("#logoutModal").css("display", "none");
    });
});

// Registration
$(document).ready(function() {
    $("#registerPopup").click(function(event) {
        event.preventDefault();

        if (typeof window.jwt === "undefined") {
            $("#registerMessage").text("Not logged in");
        } else {
            $("#registerMessage").text("Logged in");
        }

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
            $("#registerMessage").text("Registration failure: " + resp.status);
        }).done(function(resp) {
            $("#registerMessage").text("Registration success");
        });
    });

    $("#registerClose").click(function (event) {
        event.preventDefault();
        $("#registerModal").css("display", "none");
    });
});

// Preferences
$(document).ready(function() {
    $("#prefsPopup").click(function (event) {
        event.preventDefault();

        // Clear the current prefs content
        if (typeof window.jwt === "undefined") {
            $("#prefsMessage").text("Not logged in");
        } else {
            $("#prefsMessage").text("Logged in");
        }
        $("#prefsLanguages").find("option").remove();

        $.ajax({
            type: "GET",
            url: toAppPath("messages/greetings"),
            dataType: "json",
            headers: {
                "Authorization": "Bearer " + btoa(window.jwt)
            },
            data: ""
        }).fail(function(resp) {
            $("#prefsLanguages").append(
                $("<option></option>")
                .attr("value", "none")
                .text("Failed to get languages: " + resp.status));
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
            type: "POST",
            url: toAppPath("messages/user"),
            dataType: "json",
            headers: {
                "Authorization": "Bearer " + btoa(window.jwt)
            },
            data: JSON.stringify({
                user_language: language
            })
        }).fail(function(resp) {
            $("#prefsMessage").text("Failed to save preferences: " + resp.status);
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
    $("#greetPopup").click(function (event) {
        event.preventDefault();

        // Clear the current greeter UI content
        $("#greetMessage").text("No greeting");
        
        delete window.user_name;
        delete window.user_language;

        $.ajax({
            type: "GET",
            url: toAppPath("messages/user"),
            dataType: "json",
            headers: {
                "Authorization": "Bearer " + btoa(window.jwt)
            },
            data: ""
        }).fail(function(resp) {
            $("#greetMessage").text("Failed to retrieve preferences: " + resp.status);
        }).done(function(resp) {
            window.user_name = resp.user_name;
            window.user_language = resp.user_language;
            $("#greetMessage").text("Preferences of " + resp.user_name + ": " + resp.user_language);
        }).always(function() {
            $("#greetModal").css("display", "block");
        });
    });

    $("#greetBox").submit(function (event) {
        event.preventDefault();

        $.ajax({
            type: "GET",
            url: toAppPath("messages/greetings/" + window.user_language),
            dataType: "json",
            headers: {
                "Authorization": "Bearer " + btoa(window.jwt)
            },
            data: ""
        }).fail(function(resp) {
            $("#greetMessage").text("Greeting Failure: " + resp.status);
        }).done(function(resp) {
            $("#greetMessage").text("Greeting Success: " + resp.message);
        });
    });

    $("#greetClose").click(function (event) {
        event.preventDefault();
        $("#greetModal").css("display", "none");
    });
});
