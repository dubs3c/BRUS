# BRUS

    bbbbbb   rrrrrr   u     u   sssss
    b    b   r    r   u     u   s
    bbbbbb   rrrrrr   u     u   sssss
    b    b   r r      u     u       s
    bbbbbb   r  r     uuuuuuu   sssss


BRUS _(Noise in Swedish)_ parses your web server (e.g. nginx) log files and checks with GreyNoise how much noise your website is exposed to. The result can then be sent to your email address or preferred webhook service, such as Slack, Telegram and so on.

The idea is to set a cron/scheduled job that executes BRUS which will then gather log files during the last X days. Now you can get a fine summary each month for example.

_Still in active development, use at your own risk_