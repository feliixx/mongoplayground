function indent(e,n){var a="",t=!1,r=!1,c=0,s=0,i=e.charAt(s);for(e.startsWith("db.")&&(s=e.indexOf("(")+1,a+=e.substring(0,s));s<e.length;)if(" "!==(i=e.charAt(s))&&"\n"!==i&&"\t"!==i){switch(t&&"]"!==i&&"}"!==i&&(t=!1,c++,a+=n===indentMode?newline(c):""),i){case"(":r=!0,a+=i;break;case")":r=!1,a+=i;break;case"{":case"[":t=!0,a+=i;break;case",":a+=i,n===indentMode&&(a+=r?" ":newline(c));break;case":":a+=i,n===indentMode&&(a+=" ");break;case"}":case"]":t?t=!1:(c--,a+=n===indentMode?newline(c):""),a+=i;break;case'"':case"'":var d=i;for(a+='"',s++,i=e.charAt(s);i!==d&&s<e.length;)a+=i,s++,i=e.charAt(s);a+='"';break;case"n":var h=e.substring(s,s+9);if("new Date("===h){for(a+=h,s+=h.length+1,i=e.charAt(s);")"!==i&&s<e.length;)a+=i,s++,i=e.charAt(s);a+=")"}else a+=i;break;case"/":for(a+=i,s++,i=e.charAt(s);"/"!==i&&s<e.length;)a+=i,s++,i=e.charAt(s);a+="/";break;default:a+=i}s++}else s++;return a}function newline(e){for(var n="\n",a=0;a<e;a++)n+="  ";return n}const compactMode=0,indentMode=1;