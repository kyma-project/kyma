// This file is used to trigger the custom rule that checks if all markdown headings (words longer than 4 characters) are written in the title case. To run this check, you must include the check in the markdownlint command. 
// For example, if you want to run the check on the `docs` folder, run the following command: `markdownlint -r ./markdown_heading_capitalization.js docs/`.
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