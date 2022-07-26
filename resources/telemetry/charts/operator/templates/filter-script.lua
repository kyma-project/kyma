function kubernetes_map_keys(tag, timestamp, record)
    if record.kubernetes == nil then
        return 0
    end
    map_keys(record.kubernetes.annotations)
    map_keys(record.kubernetes.labels)
    return 1, timestamp, record
end
function map_keys(table)
    if table == nil then
        return
    end
    local new_table = {}
    local changed_keys = {}
    for key, val in pairs(table) do
        local mapped_key = string.gsub(key, "[%/%.]", "_")
        if mapped_key ~= key then
            new_table[mapped_key] = val
            changed_keys[key] = true
        end
    end
    for key in pairs(changed_keys) do
        table[key] = nil
    end
    for key, val in pairs(new_table) do
        table[key] = val
    end
end