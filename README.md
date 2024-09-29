Small tool to perform SCCM recon and enumerate a Primary Site Server (PSS) or Distrubution Point (DP) over winreg. This can enumerate useful information such as the Site Database, whether a DP allows anonymous access, if a DP is PXE enabled and the location of Management Points (MP) in the site.

Usage:
```
pssrecon -u <USER> -p <PASS> -d <DOMAIN> -host <PRIMARY SITE SERVER>
```
Example:
```
pssrecon -u lowpriv -p password -d corp.local -host pss.corp.local

[+] Distrubution Point Installed
[+] Site Code Found: COR
[+] Site Server Found: SCCM.corp.local
[+] Management Point Found: http://SCCM.corp.local
[+] Management Point Found: http://SCCMMP.corp.local
[+] Management Point Installed
[+] Site Database Found: SCCMDB01.CORP.LOCAL

pssrecon -u lowpriv -p password -d corp.local -host dp.corp.local

[+] Distrubution Point Installed
[+] Site Code Found: COR
[+] Site Server Found: SCCM.CORP.local
[+] Management Point Found: http://SCCM.corp.local
[+] Management Point Found: http://SCCMMP.corp.local
[+] Anonymous Access On This Distrubution Point Is Enabled
[+] PXE Installed

```
