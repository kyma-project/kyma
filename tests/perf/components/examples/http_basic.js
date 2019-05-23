import http from 'k6/http';

export default function() {
    const response = http.get("https://www.google.com");
}