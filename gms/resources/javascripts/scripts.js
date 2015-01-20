var tags=[];
var tagNo=0;
var tagTest=3;
$(document).ready(function() {

	$("#imgInp").change(function(){
		console.log("=====================================");
	    readURL(this);
		document.getElementById('submit').style.visibility = 'visible';
	});
	
	$("#uploadForm").keypress(function(e) {
		  //Enter key
		if (e.which == 13) {
		    return false;
		}
	});
	
	$("#enterTag").unbind('keypress').keypress(function(e) {
		  //Enter key
		if (e.which == 13) {
			var tag = $("input#enterTag").val();
			console.log("entering tag");
			console.log(tag);
			
			if (tag != "" && tag != " ") {
				addTag(tag, "displayTags");
				if (document.getElementById('displayTags').style.visibility == 'hidden'){
					document.getElementById('displayTags').style.visibility = 'visible';
				}
			
				flickrRelatedTags(tag);
			}
			$('#enterTag').val("");
		}
	});
});


function addTag(tag, tagDiv) {
	var x = document.getElementById(tagDiv);
	var option = document.createElement("a");
	var tagId = "tag"+tagNo++;
				
	option.text = tag;
	option.setAttribute('id',tagId);
	
	if (tagDiv == "displayTags"){
		option.setAttribute('class', "tagUpload");
		option.setAttribute('onClick', function(event){removeTag(tagDiv);});
		option.onclick = function() {removeTag(tagDiv);};
		tags.push(tag);
		updateTagList();
	} else if (tagDiv == "suggestedTags") {
		option.setAttribute('class', "suggestedTag");
		option.setAttribute('onClick', function(event){addToMainList(tag);});
		option.onclick = function() {addToMainList(tag);};
		
		
	}
	x.appendChild(option);
}

function removeTag(list) {
	console.log("in remove");
	console.log(list);
	var text = $(event.target).text();
	var index = jQuery.inArray(text,tags);
	var tagList = document.getElementById(list);
	var tag = document.getElementById(event.target.id);
	tagList.removeChild(tag);
	if (index != -1) {
		console.log(event.target.id);	
		console.log(index);
		tags.splice(index, 1);
		console.log(tags);	
		updateTagList();
	}
}

function addToMainList(tag) {
	var x = document.getElementById("displayTags");
	var option = document.createElement("a");
	var tagId = "tag"+tagNo++;
	removeTag("suggestedTags");
				
	option.text = tag;
	option.setAttribute('id',tagId);
	option.setAttribute('class', "tagUpload");
	option.setAttribute('onClick', function(event){removeTag("displayTags");});
	option.onclick = function() {removeTag("displayTags");};
	tags.push(tag);
	updateTagList();
	x.appendChild(option);
}

function updateTagList() {
	var tagsForHTML = document.getElementById("tagList");
	tagsForHTML.setAttribute('value', tags);
}

function flickrRelatedTags(tag) {
	var url = "https://api.flickr.com/services/rest/?method=flickr.tags.getRelated&api_key=ef72e911f885e924a460b98a4801ff14&tag="+tag+"&per_page=5&format=json"
	$.ajax({
        url: "/flickr?" + url,
        type: "GET",
        success: function (data) {
			var myNode = document.getElementById("suggestedTags");
			while (myNode.firstChild) {
   				 myNode.removeChild(myNode.firstChild);
			}

            processFlickrTags(data)
        },
            error: function(data) {
                console.log("Error flickr");
                var err = ("(" + xhr.responseText + ")");
                console.log(data);
            }
    });
}

function processFlickrTags(tags) {
	var indivTags = tags.split(',');
	indivTags.pop();

	console.log(indivTags);
	for (i=0; i<indivTags.length; i++) {
		addTag(indivTags[i], "suggestedTags")
	}
	if (document.getElementById('suggestedTags').style.visibility == 'hidden'){
		document.getElementById('suggestedTags').style.visibility = 'visible';
	}
}




function readURL(input) {
		console.log("///////////////////////////////////////")
		console.log(input);
        if (input.files && input.files[0]) {
            var reader = new FileReader();
            
            reader.onload = function (e) {
                $('#blah').attr('src', e.target.result);
				document.getElementById('blah').style.visibility='visible';
				document.getElementById('photoDetails').style.visibility='visible';
				console.log(e.target.result);
            }
            
            reader.readAsDataURL(input.files[0]);
        }
    }
    
function tagCloud() {
	var tagMap = {};
	$.ajax({
        url: "/tagCloud",
        type: "GET",
        success: function (data) {
			console.log(data);
			var t=data.split(',');
			var max = parseInt(t.pop().split(' ')[1]);
			console.log(max);
			for (i=0; i<t.length; i++) {
				var split=t[i].split(' ');
				tagMap[split[0]]=parseInt(split[1]);
			}
			console.log(tagMap);
			for (var m in tagMap){
				if(tagMap[m] > 0){
					console.log(tagMap+"***************************");
					console.log(max/tagMap[m]+"max");
					if(tagMap[m]/max == 1) size = 8;
					else if(tagMap[m]/max == 0.5) size = 6;
					else if ((0.3<tagMap[m]/max) && (tagMap[m]/max<0.5)) size = 4;
					else size = 2;
					$('#cloud').append("<a class='size-"+size+"' href='/tag?"+m+"'>"+m+"</a>");
				}
			}
			
        },
            error: function(data) {
                console.log("Error getting tags from db");
            }
    });
	
	
	
}