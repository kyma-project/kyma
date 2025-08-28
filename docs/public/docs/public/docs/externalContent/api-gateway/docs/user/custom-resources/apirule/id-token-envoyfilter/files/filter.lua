function envoy_on_request(request_handle)
    headers = request_handle:headers()
    test_header = headers:get("X-Forwarded-Host")
    new_header = string.format("%s:%s", test_header, "generated_value")
    request_handle:headers():replace("X-Forwarded-Host", new_header)
end
function envoy_on_response(response_handle)
  --
end