# Bytecode Installation Instructions

As a set of tools designed to deal with blockchain data, please be aware that we're talking about a lot of the data. Later on I will give approximate insight into how much space you will need to squeze full potential of this repository.

Initial idea was to have data directly placed under the `data/` path, however, some of the data, such
as contracts, signatures can be quite large and therefore, these datasets you will need to obtain manually.

Luckly, all you need is this document and patience as for some of the datasets you'll need to leave machine over night to handle processing. 

## Command Overview

### BSC Scan Verified Contracts

Bscscan has neat [BscScan Verified Contracts CSV] page from which you can download all of the verified
contracts that contain corresponding open source license. This is assumptions they are all, I am not 100% sure that is the case.

Additionaly, there are contracts available through [BscScan Contracts API] that are not in the CSV file.

Daily verified contracts can be seen at [BscScan Daily Verified Contracts].

Following command will look into `/data/bscscan/verified-contracts.csv` and fetch all of contracts
in the CSV into `/data/bscscan/verified-contracts.gob` for future processing. On that subject I will speak later.

**WARNINGS:**

- You can download contract information with free API key but probably you will need bscscan subscription to get all of these data smoothly and not to be heavily rate limited. Rate limit 1st paid subscription level is small as well (10/s).
- You will need manually to download CSV files in time to time. I will be doing it as well but if I miss it, well you'll have to do it and replace files. There is no automatic way to download the csv as it's behind the captcha.

```
txbyte syncer bscscan --config ./.txbyte.yaml
```

#### Contract Processing

To be defined...

### FourBytes Database Download

[4byte.dictionary] is a free service from which you can find few million method signatures from a lot of the contracts. In order to reverse engineer functions from the contracts that did not publish their source code and ABI, service like this helps a lot as we are able to at least partially discover methods including arguments. Arguments are not named tho, but they are arguments, none the less.

This is free service, and one of requirements and what they pleased consumers is to use API with care. Therefore, running this command will take few hours to download whole database.

Great thing is that we are keeping processed pages so if you stop it, or restart, it will continue from the page it left.

Data is stored in [BadgerDB].

```
txbyte syncer fourbyte --config ./.txbyte.yaml
```



[4byte.dictionary]: <https://www.4byte.directory/>
[BscScan Daily Verified Contracts]: <https://bscscan.com/contractsVerified>
[BscScan Verified Contracts CSV]: <https://bscscan.com/exportData?type=open-source-contract-codes>
[BscScan Contracts API]: <https://docs.bscscan.com/api-endpoints/contracts>
[BadgerDB]: <https://github.com/dgraph-io/badger>