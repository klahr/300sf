#!/usr/bin/bash

300sf zone --insert Sweden --admin_level country --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/sverige.geojson

ID=$(300sf zone --lang en | grep Sweden | grep country | awk '{print $1}')

300sf zone --insert "Värmland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/varmland.geojson --parent_zone_id="$ID"
300sf zone --insert "Östergötland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/ostergotland.geojson --parent_zone_id="$ID"
300sf zone --insert "Örebro County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/orebro.geojson --parent_zone_id="$ID"
300sf zone --insert "Västra Götaland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/vastra-gotaland.geojson --parent_zone_id="$ID"
300sf zone --insert "Västmanland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/vastmanland.geojson --parent_zone_id="$ID"
300sf zone --insert "Västernorrland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/vasternorrland.geojson --parent_zone_id="$ID"
300sf zone --insert "Västerbotten County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/vasterbotten.geojson --parent_zone_id="$ID"
300sf zone --insert "Uppsala County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/uppsala.geojson --parent_zone_id="$ID"
300sf zone --insert "Södermanland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/sodermanland.geojson --parent_zone_id="$ID"
300sf zone --insert "Stockholm County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/stockholm.geojson --parent_zone_id="$ID"

# Skåne
ID=$(300sf zone --insert "Skåne County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/skane.geojson --parent_zone_id="$ID")
300sf zone --insert "Östra Göinge kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/ostra-goinge.geojson --parent_zone_id="$ID"
300sf zone --insert "Bjuvs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/bjuv.geojson --parent_zone_id="$ID"
300sf zone --insert "Bromölla kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/bromolla.geojson --parent_zone_id="$ID"
300sf zone --insert "Burlövs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/bulov.geojson --parent_zone_id="$ID"
300sf zone --insert "Båstads kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/bastad.geojson --parent_zone_id="$ID"
300sf zone --insert "Eslövs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/eslov.geojson --parent_zone_id="$ID"
300sf zone --insert "Helsingborgs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/helsingborg.geojson --parent_zone_id="$ID"
300sf zone --insert "Hässleholms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/hassleholm.geojson --parent_zone_id="$ID"
300sf zone --insert "Höganäs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/hoganas.geojson --parent_zone_id="$ID"
300sf zone --insert "Hörby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/horby.geojson --parent_zone_id="$ID"
300sf zone --insert "Höörs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/hoor.geojson --parent_zone_id="$ID"
300sf zone --insert "Klippans kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/klippans.geojson --parent_zone_id="$ID"
300sf zone --insert "Kristianstads kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/kavlinge.geojson --parent_zone_id="$ID"
300sf zone --insert "Kävlinge kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/kavlinge.geojson --parent_zone_id="$ID"
300sf zone --insert "Landskrona kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/landskrona.geojson --parent_zone_id="$ID"
300sf zone --insert "Lomma kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/lomma.geojson --parent_zone_id="$ID"
300sf zone --insert "Lunds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/lund.geojson --parent_zone_id="$ID"
300sf zone --insert "Malmö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/malmo.geojson --parent_zone_id="$ID"
300sf zone --insert "Osby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/osby.geojson --parent_zone_id="$ID"
300sf zone --insert "Perstorps kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/perstorp.geojson --parent_zone_id="$ID"
300sf zone --insert "Simrishamns kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/simrishamn.geojson --parent_zone_id="$ID"
300sf zone --insert "Sjöbo kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/sjobo.geojson --parent_zone_id="$ID"
300sf zone --insert "Skurups kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/skurup.geojson --parent_zone_id="$ID"
300sf zone --insert "Staffanstorps kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/staffanstorp.geojson --parent_zone_id="$ID"
300sf zone --insert "Svalövs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/svalovs.geojson --parent_zone_id="$ID"
300sf zone --insert "Svedala kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/svedala.geojson --parent_zone_id="$ID"
300sf zone --insert "Tomelilla kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/tomelilla.geojson --parent_zone_id="$ID"
300sf zone --insert "Trelleborgs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/trelleborg.geojson --parent_zone_id="$ID"
300sf zone --insert "Vellinge kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/vellinge.geojson --parent_zone_id="$ID"
300sf zone --insert "Ystads kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/ystad.geojson --parent_zone_id="$ID"
300sf zone --insert "Ängelholms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/angelholm.geojson --parent_zone_id="$ID"
300sf zone --insert "Åstorps kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/astorp.geojson --parent_zone_id="$ID"
300sf zone --insert "Örkelljunga kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/skane/orkelljunga.geojson --parent_zone_id="$ID"

# Norrbotten
ID=$(300sf zone --insert "Norrbotten County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/norrbotten.geojson --parent_zone_id="$ID")
300sf zone --insert "Övertorneå kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/overtornea.geojson --parent_zone_id="$ID"
300sf zone --insert "Arjeplogs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/arjeplog.geojson --parent_zone_id="$ID"
300sf zone --insert "Arvidsjaurs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/arvidsjaur.geojson --parent_zone_id="$ID"
300sf zone --insert "Bodens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/boden.geojson --parent_zone_id="$ID"
300sf zone --insert "Gällivare kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/gallivare.geojson --parent_zone_id="$ID"
300sf zone --insert "Haparanda kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/haparanda.geojson --parent_zone_id="$ID"
300sf zone --insert "Jokkmokks kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/jokkmokk.geojson --parent_zone_id="$ID"
300sf zone --insert "Kalix kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/kalix.geojson --parent_zone_id="$ID"
300sf zone --insert "Kiruna kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/kiruna.geojson --parent_zone_id="$ID"
300sf zone --insert "Luleå kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/lulea.geojson --parent_zone_id="$ID"
300sf zone --insert "Pajala kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/pajala.geojson --parent_zone_id="$ID"
300sf zone --insert "Piteå kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/pitea.geojson --parent_zone_id="$ID"
300sf zone --insert "Älvsbyns kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/alvsbyn.geojson --parent_zone_id="$ID"
300sf zone --insert "Överkalix kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/norrbotten/overkalix.geojson --parent_zone_id="$ID"

# Kronoberg
ID=$(300sf zone --insert "Kronoberg County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/kronoberg.geojson --parent_zone_id="$ID")
300sf zone --insert "Älmhults kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/almhult.geojson --parent_zone_id="$ID"
300sf zone --insert "Växjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/vaxjo.geojson --parent_zone_id="$ID"
300sf zone --insert "Uppvidinge kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/uppvidinge.geojson --parent_zone_id="$ID"
300sf zone --insert "Tingsryds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/tingsryd.geojson --parent_zone_id="$ID"
300sf zone --insert "Markaryds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/markaryd.geojson --parent_zone_id="$ID"
300sf zone --insert "Ljungby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/ljungby.geojson --parent_zone_id="$ID"
300sf zone --insert "Alvesta kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/alvesta.geojson --parent_zone_id="$ID"
300sf zone --insert "Lessebo kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kronoberg/lessebo.geojson --parent_zone_id="$ID"

# Kalmar
ID=$(300sf zone --insert "Kalmar County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/kalmar.geojson --parent_zone_id="$ID")
300sf zone --insert "Västerviks kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/vasterviks.geojson --parent_zone_id="$ID"
300sf zone --insert "Borgholms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/borgholm.geojson --parent_zone_id="$ID"
300sf zone --insert "Emmaboda kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/emmaboda.geojson --parent_zone_id="$ID"
300sf zone --insert "Hultsfreds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/hultsfred.geojson --parent_zone_id="$ID"
300sf zone --insert "Högsby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/hogsby.geojson --parent_zone_id="$ID"
300sf zone --insert "Kalmar kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/kalmars.geojson --parent_zone_id="$ID"
300sf zone --insert "Mönsterås kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/monsteras.geojson --parent_zone_id="$ID"
300sf zone --insert "Mörbylånga kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/morbylanga.geojson --parent_zone_id="$ID"
300sf zone --insert "Nybro kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/nybro.geojson --parent_zone_id="$ID"
300sf zone --insert "Oskarshamns kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/oskarshamn.geojson --parent_zone_id="$ID"
300sf zone --insert "Torsås kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/torsas.geojson --parent_zone_id="$ID"
300sf zone --insert "Vimmerby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/kalmar/vimmerby.geojson --parent_zone_id="$ID"

# Jönköping
ID=$(300sf zone --insert "Jönköping County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/jonkoping.geojson --parent_zone_id="$ID")
300sf zone --insert "Värnamo kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/varnamo.geojson --parent_zone_id="$ID"
300sf zone --insert "Vetlanda kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/vetlanda.geojson --parent_zone_id="$ID"
300sf zone --insert "Vaggeryds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/vaggeryd.geojson --parent_zone_id="$ID"
300sf zone --insert "Tranås kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/tranas.geojson --parent_zone_id="$ID"
300sf zone --insert "Sävsjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/savsjo.geojson --parent_zone_id="$ID"
300sf zone --insert "Nässjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/nassjo.geojson --parent_zone_id="$ID"
300sf zone --insert "Mullsjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/mullsjo.geojson --parent_zone_id="$ID"
300sf zone --insert "Jönköpings kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/jonkopings.geojson --parent_zone_id="$ID"
300sf zone --insert "Aneby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/aneby.geojson --parent_zone_id="$ID"
300sf zone --insert "Eksjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/eksjö.geojson --parent_zone_id="$ID"
300sf zone --insert "Gislaveds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/gislaved.geojson --parent_zone_id="$ID"
300sf zone --insert "Gnosjö kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/gnosjo.geojson --parent_zone_id="$ID"
300sf zone --insert "Habo kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jonkoping/habo.geojson --parent_zone_id="$ID"

# Jämtland
ID=$(300sf zone --insert "Jämtland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/jamtland.geojson --parent_zone_id="$ID")
300sf zone --insert "Östersunds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/ostersund.geojson --parent_zone_id="$ID"
300sf zone --insert "Bergs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/berg.geojson --parent_zone_id="$ID"
300sf zone --insert "Bräcke kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/bracke.geojson --parent_zone_id="$ID"
300sf zone --insert "Härjedalens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/harjedalen.geojson --parent_zone_id="$ID"
300sf zone --insert "Krokoms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/krokoms.geojson --parent_zone_id="$ID"
300sf zone --insert "Ragunda kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/ragunda.geojson --parent_zone_id="$ID"
300sf zone --insert "Strömsunds kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/stromsund.geojson --parent_zone_id="$ID"
300sf zone --insert "Åre kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/jamtland/are.geojson --parent_zone_id="$ID"

# Halland
ID=$(300sf zone --insert "Halland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/halland.geojson --parent_zone_id="$ID")
300sf zone --insert "Varbergs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/varberg.geojson --parent_zone_id="$ID"
300sf zone --insert "Laholms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/laholm.geojson --parent_zone_id="$ID"
300sf zone --insert "Kungsbacka kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/kungsbacka.geojson --parent_zone_id="$ID"
300sf zone --insert "Hylte kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/hylte.geojson --parent_zone_id="$ID"
300sf zone --insert "Halmstads kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/halmstad.geojson --parent_zone_id="$ID"
300sf zone --insert "Falkenbergs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/halland/falkenberg.geojson --parent_zone_id="$ID"

# Gävleborg
ID=$(300sf zone --insert "Gävleborg County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/gavleborg.geojson --parent_zone_id="$ID")
300sf zone --insert "Söderhamns kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/soderhamns.geojson --parent_zone_id="$ID"
300sf zone --insert "Bollnäs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/bollnas.geojson --parent_zone_id="$ID"
300sf zone --insert "Gävle kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/gavle.geojson --parent_zone_id="$ID"
300sf zone --insert "Hofors kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/hofors.geojson --parent_zone_id="$ID"
300sf zone --insert "Hudiksvalls kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/hudiksvall.geojson --parent_zone_id="$ID"
300sf zone --insert "Ljusdals kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/ljusdal.geojson --parent_zone_id="$ID"
300sf zone --insert "Nordanstigs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/nordanstig.geojson --parent_zone_id="$ID"
300sf zone --insert "Ockelbo kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/ockelbo.geojson --parent_zone_id="$ID"
300sf zone --insert "Ovanåkers kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/ovanaker.geojson --parent_zone_id="$ID"
300sf zone --insert "Sandvikens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gavleborg/sandviken.geojson --parent_zone_id="$ID"

# Gotland
ID=$(300sf zone --insert "Gotland County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/gotland.geojson --parent_zone_id="$ID")
300sf zone --insert "Gotlands kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/gotland/gotlands.geojson --parent_zone_id="$ID"

# Dalarna
ID=$(300sf zone --insert "Dalarna County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/dalarna.geojson --parent_zone_id="$ID")
300sf zone --insert "Älvdalens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/alvdalen.geojson --parent_zone_id="$ID"
300sf zone --insert "Avesta kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/avesta.geojson --parent_zone_id="$ID"
300sf zone --insert "Borlänge kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/borlange.geojson --parent_zone_id="$ID"
300sf zone --insert "Falu kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/falun.geojson --parent_zone_id="$ID"
300sf zone --insert "Gagnefs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/gagnef.geojson --parent_zone_id="$ID"
300sf zone --insert "Hedemora kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/hedemora.geojson --parent_zone_id="$ID"
300sf zone --insert "Leksands kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/leksand.geojson --parent_zone_id="$ID"
300sf zone --insert "Ludvika kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/ludvika.geojson --parent_zone_id="$ID"
300sf zone --insert "Malung-Sälens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/malung.geojson --parent_zone_id="$ID"
300sf zone --insert "Mora kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/mora.geojson --parent_zone_id="$ID"
300sf zone --insert "Orsa kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/orsa.geojson --parent_zone_id="$ID"
300sf zone --insert "Rättviks kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/rattvik.geojson --parent_zone_id="$ID"
300sf zone --insert "Smedjebackens kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/smedjebacken.geojson --parent_zone_id="$ID"
300sf zone --insert "Säters kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/sater.geojson --parent_zone_id="$ID"
300sf zone --insert "Vansbro kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/dalarna/vansbro.geojson --parent_zone_id="$ID"

# Blekinge
ID=$(300sf zone --insert "Blekinge County" --admin_level county --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/lan/blekinge.geojson --parent_zone_id="$ID")
300sf zone --insert "Sölvesborgs kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/blekinge/solvesborg.geojson --parent_zone_id="$ID"
300sf zone --insert "Ronneby kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/blekinge/ronneby.geojson --parent_zone_id="$ID"
300sf zone --insert "Olofströms kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/blekinge/olofstroms.geojson --parent_zone_id="$ID"
300sf zone --insert "Karlskrona kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/blekinge/karlskrona.geojson --parent_zone_id="$ID"
300sf zone --insert "Karlshamns kommun" --admin_level municipality --polygon /home/klarre/dev/sourcehut/~klahr/300sf/boundries/sverige/blekinge/karlshamn.geojson --parent_zone_id="$ID"

