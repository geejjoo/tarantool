box.cfg{
    listen = 3301,
    log_level = 5,
}

-- Создать пользователя, если не существует
local user = 'admin'
local password = 'admin'
if not box.schema.user.exists(user) then
    box.schema.user.create(user, { password = password })
    box.schema.user.grant(user, 'super', nil, nil)
else
    -- Если пользователь уже есть, можно обновить пароль (опционально)
    box.schema.user.passwd(user, password)
end

-- Создать пространство, если не существует
if not box.space.kv then
    box.schema.space.create('kv', { if_not_exists = true })
    box.space.kv:format({
        { name = 'key', type = 'string' },
        { name = 'value', type = '*' },
        { name = 'created_at', type = 'unsigned' },
        { name = 'updated_at', type = 'unsigned' },
        { name = 'deleted_at', type = 'unsigned' },
        { name = 'is_deleted', type = 'boolean' },
    })
    box.space.kv:create_index('primary', { parts = { 'key',  }, if_not_exists = true })
    box.space.kv:create_index('deleted', { parts = { 'is_deleted', 'key' }, if_not_exists = true })
end

-- Всё готово, можно принимать соединения
print('Tarantool minimal init complete')