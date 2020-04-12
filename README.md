# Link Shortener
This is a simple proof of concept for a temporary link shortener. More to be added here later

### To Use
Just clone everything into a folder and run "main.go". The program starts a server on port 8080 which can be used as required. All links are stored in "links.db" which is created if it doesn't already exist. 

There also is the option to run this program as a docker container. I still need to set it up to accept a volume for database storage but other then that, there should be no issue running in a container. Note that this does build using via multi-stage images to keep final build size small but that does mean you may have some random images remaining in addition to the final image. To remove them simply use `docker image prune --filter label=stage=builder`. Please let me know if there are issues as this is one of my first times using Docker :).


Any pull requests and issues are welcome! This is a super simple skill builder side project so I may not fix everthing but I have found it suprisingly useful to have this as a service