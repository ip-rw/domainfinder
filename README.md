# domain finder
this package is all about finding domain names for organizations given only a limited amount of information, potentially 
just a name.

the way this works is effectively:
1. (to obtain test data), the finder tool expects a company number as the first argument, we will look this up on opencorporates.com 
   and make a company object made up of any information that will be useful when we come to try and match a domain
2. generate candidate domains for our company object. currently implemented are:
    * duckduckgo
    * guesswork
    * clearbit
3. we crawl the (hopefully) websites on the canditate domains, collecting all the text we find there.
4. use magic (latent semantic analysis) to find the website most similar to data contained within our company object.  this 
   comes in the form of score between 0 and 1.
   
The code in cmd/finder is ugly but there's no point of rewriting what is basically just a demo.