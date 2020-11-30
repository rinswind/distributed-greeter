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
            }).done(function(resp) {
                $("#loginMessage").text("Logout success");
            });
    
            $("#logoutMessage").text("Logged out");
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

// Greeting
$(document).ready(function() {
    $("#greetPopup").click(function (event) {
        event.preventDefault();

        // Clear the current greeter UI content
        $("#greetMessage").text("No greeting");
        $("#greetLanguages").find("option").remove();

        $.ajax({
            type: "GET",
            url: toAppPath("messages/greetings"),
            dataType: "json",
            headers: {
                // Retrieve login from the window object
                "Authorization": "Bearer " + btoa(window.jwt)
            },
            data: ""
        }).fail(function(resp) {
            $("#greetLanguages").append(
                $("<option></option>")
                .attr("value", "none")
                .text("Failed to get languages: " + resp.status));
        }).done(function(resp) {
            $.each(resp.languages, function(key, val) {
                $("#greetLanguages").append(
                    $("<option></option>")
                    .attr("value", key)
                    .text(key));
            });
        }).always(function() {
            $("#greetModal").css("display", "block");
        });
    });

    $("#greetBox").submit(function (event) {
        event.preventDefault();

        var language = $("#greetLanguages").find(":selected").text();
        
        $.ajax({
            type: "GET",
            url: toAppPath("messages/greetings/" + language),
            dataType: "json",
            headers: {
                // Retrieve login from the window object
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
