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

```
➜  BRUS git:(master) ✗ ./brus -webhook -directory "/var/log/nginx/"
🚀 Data sent to webhook
```

## Contributing
1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D