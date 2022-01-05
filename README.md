# BRUS

    bbbbb    rrrrrr   u     u   sssss
    b    b   r    r   u     u   s
    bbbbb    rrrrrr   u     u   sssss
    b    b   r r      u     u       s
    bbbbb    r  r     uuuuuuu   sssss


BRUS _(Noise in Swedish)_ parses your web server (e.g. nginx) log files and checks with GreyNoise how much noise your website is exposed to. The result can then be sent to your email address or preferred webhook service, such as Slack, Telegram and so on.

The idea is to set a cron/scheduled job that executes BRUS which will then gather log files during the last X days. Now you can get a fine summary each month for example.

_Still in active development, use at your own risk_

## Usage

Create the following config file in `~/.config/brus.ini`
```
[Email]
username=
password=
recipient=
server=smtp.gmail.com
port=587
subject="BRUS summary"

[Webhook]
webhook=https://api.telegram.org/botxxxxxx:xxxxxxxxxxxxxxxxxxxxxxxxxxxxx/sendMessage
textField=text
data={"chat_id": "xxxxxx"}

[GreyNoise]
key=
```

The `textField` and `data` fields are used for telegram because they require some extra fields. If your webhook request uses `message=your+data` in its POST payload, then you only need the webhook field. If it uses another name, such as `text=your+data`, you can change it in `textField`. Do you need to send more data, simply add it as a json formatted string in the `data` field.

Now run the program :)
```
➜  BRUS git:(master) ✗ ./brus -webhook -directory "/var/log/nginx/"
🚀 Data sent to webhook

# Results from BRUS the last 30 days
Amount of Noisy IPs: 15
Non Noisy IPs: 1
Top 3 Classification: unknown, malicious
Top 3 Names: unknown, Net Systems Research
```


## Contributing
1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D