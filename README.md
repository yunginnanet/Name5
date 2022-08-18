# name5

the main DNS code that powers professorOak's aggressive latteral resolver

##### `go run ./ aol.com`

```
---------name-
name: aol.com.
212.82.100.150
124.108.115.100
98.136.103.23
74.6.136.150
.,.,.
name: w2.src.vip.tw1.yahoo.com.
2406:2000:fc:c5f::a000
.,.,.
name: w2.src.vip.ir2.yahoo.com.
2a00:1288:110:c305::1:9000
.,.,.
name: w2.src.vip.bf1.yahoo.com.
2001:4998:124:1507::4000
.,.,.
name: w2.src.vip.sg3.yahoo.com.
106.10.248.150
2406:2000:e4:1605::4000
.,.,.
name: w2.src.vip.gq1.yahoo.com.
2001:4998:24:120d::5000
-------------


----------ptr-
ip: 2406:2000:e4:1605::4000 
ptr: w2.src.vip.sg3.yahoo.com.
+ - + - +
ip: 2001:4998:124:1507::4000 
ptr: w2.src.vip.bf1.yahoo.com.
+ - + - +
ip: 2a00:1288:110:c305::1:9000 
ptr: w2.src.vip.ir2.yahoo.com.
+ - + - +
ip: 106.10.248.150 
ptr: w2.src.vip.sg3.yahoo.com.
+ - + - +
ip: 124.108.115.100 
ptr: w2.src.vip.tw1.yahoo.com.
+ - + - +
ip: 74.6.136.150 
ptr: w2.src.vip.bf1.yahoo.com.
+ - + - +
ip: 98.136.103.23 
ptr: w2.src.vip.gq1.yahoo.com.
+ - + - +
ip: 212.82.100.150 
ptr: w2.src.vip.ir2.yahoo.com.
+ - + - +
ip: 2406:2000:fc:c5f::a000 
ptr: w2.src.vip.tw1.yahoo.com.
+ - + - +
ip: 2001:4998:24:120d::5000 
ptr: w2.src.vip.gq1.yahoo.com.
-------------
```
