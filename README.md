I made this to learn a little bit about golang. It Outputs ASCII art showing the suns approximate current position

                                          *   *              
                                      *    + +    *          
                           .     -      + .---. +            
                     .              * +  /     \ + *         
                                    + *  \     / * +         
               .                        + '---' +   .        
                                      *    + +    *          
          .                               *   *          .   
                                                             
        .                                                  . 
                                                             
                                 o                           
       .                        <|\                         .
                      _______====\7===______                 
                                                             
        .                                                  . 
                                                             
          .                                              .   
                                                             
                -                                 -          
                        .                 .                  
                                 .

### To use
```
git clone https://github.com/bocsir/sunspot-cli.git
cd sunspot-cli
go run main.go
```
You'll also need to get an API key geocode.maps.co in a file named `local.env` as `GEOCODEKEY=<API_KEY>`

### To change location
```
> coords.txt
go run main.go
```
or just delete the coords in `coords.txt` and re-run

#### APIs used:
- geocode.maps.co
- api.sunrisesunset.io
