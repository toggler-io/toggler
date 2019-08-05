# Deploying to heroku

To deploy to `heroku`, try the following steps:

* go to your dashboard
* new
* Create new app -> mycompany-toggler
* `on app page`
* go to the deploy tab
* Deployment method -> Heroku Git
	* $ heroku login
	* $ heroku git:clone -a mycompany-toggler
	* $ cd mycompany-toggler
	* $ git remote add origin git@github.com:adamluzsi/toggler.git
	* $ git pull origin master
	* $ git push heroku master
