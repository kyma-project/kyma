
let accounts = ./accounts.conf as Text

in
''
port: 4222

${accounts}
''
