module.exports = [{
    "names": [ "custom/capitalize-headings" ],
    "description": "Heading words longer than 4 characters should be capitalized",
    "tags": [ "formatting" ],
    "function": function rule(params, onError) {
      params.tokens.filter(function filterToken(token) {
        return token.type === "heading_open";
      }).forEach(function forToken(heading) {
        var headingTokenContent = heading.line.trim();
        var wordsInHeading = headingTokenContent.split(' ');
  
        for (var i = 0; i < wordsInHeading.length; i++) {
          if (wordsInHeading[i].length > 4 && wordsInHeading[i] &&
            wordsInHeading[i].charAt(0) !== wordsInHeading[i].charAt(0).toUpperCase()) {
            var capitalizedWord = wordsInHeading[i].charAt(0).toUpperCase() + wordsInHeading[i].slice(1);
            var detailMessage = "Change " + "'" + wordsInHeading[i] + "'" + " to " + "'" + capitalizedWord + "'";
  
            onError({
              "lineNumber": heading.lineNumber,
              "detail": detailMessage,
              "context": headingTokenContent, // Show the whole heading as context
              "range": [headingTokenContent.indexOf(wordsInHeading[i]), wordsInHeading[i].length] // Underline the word which needs a change
            });
          }
        }
      });
    }
  }];