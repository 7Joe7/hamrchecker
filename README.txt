HamrChecker is a simple utility for checking free badminton courts on Hamr Sport Branik
(further functionality for other locations and sports may be added upon request)

Usage:

1) in command line go to the folder where you have downloaded HamrChecker
2) start HamrChecker like this: .\HamrChecker.exe -beginningTime <time in format HH:mm> -date <date in format YYYY-MM-DD> -halfHoursToSearch <how many half hours from the beginning time it is still acceptable to get the court> -reservationLength <how many half hours do you need in a row> -to <your_email_address>

NOTE: the .exe file is for Windows, the file without extension is for Linux or Mac OS

Further description of the parameters:

beginningTime - time from which you want to get the court (e.g. if I want to play from 5 pm I will put here 17:00)
date - date at which you want to get the court in format YYYY-MM-DD !! BE CAREFUL ABOUT ZEROES e.g. valid is only 2016-04-19 and NOT 2016-4-19 !! (e.g. if I want to play on 19.4.2016 I will input 2016-04-19)
halfHoursToSearch - till what time I don't mind getting the court (e.g. I want to play sometime between 17:00 and 19:00 for one hour so I will input here all the half hours between 17:00 and 19:00 which is 4 - number of half hours in that range)
reservationLength - how long reservation do I need in half hours (e.g. I want a court for an hour so I will put here 2 - two half hours)
to - to whom it should send a notification email when the court is free to be reserved (e.g. my email is josef.erneker@gmail.com so I will put here this email). This parameter supports multiple emails, separate them by comma like this: <first_email_address,second_email_address,third_email_address>.

The utility depends on having an Internet connection.
The utility will not reserve the court for you, it will just let you know when it is free.
If you want to cancel the search press ctrl + C in the running console.

Have a nice day.

Contact me if there is any problem.

Josef Erneker <josef.erneker@gmail.com>